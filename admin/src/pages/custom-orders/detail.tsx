import { useState, useEffect, useCallback } from "react";
import {
  Card, Row, Col, Tag, Button, Space, Descriptions, Table, App,
  Modal, Form, InputNumber, Input, Upload, Typography, Badge,
  Popconfirm, Divider,
} from "antd";
import {
  ArrowLeftOutlined, CheckOutlined, CloseOutlined, SendOutlined,
  DollarOutlined, EditOutlined, UploadOutlined, DeleteOutlined,
  LinkOutlined, SyncOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useNavigate, useParams } from "react-router-dom";
import api from "../../lib/api";

const { Text, Link } = Typography;
const { TextArea } = Input;

interface OrderItem {
  id: number;
  productId?: number;
  customItemName?: string;
  customItemDescription?: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
}

interface CustomDetails {
  id: number;
  clientDescription?: string;
  adminNotes?: string;
  fileUrls: string[];
  printSettings: Record<string, unknown>;
  bitrixDealId?: string;
  bitrixStageId?: string;
  createdAt: string;
  updatedAt: string;
}

interface CustomOrder {
  id: number;
  orderNumber: string;
  status: string;
  orderType: string;
  subtotal: number;
  totalPrice: number;
  isPaid: boolean;
  paymentMethod: string;
  paymentLink?: string;
  paymentExpiresAt?: string;
  deliveryMethod: string;
  deliveryAddress?: string;
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  notes?: string;
  items: OrderItem[];
  customDetails?: CustomDetails;
  createdAt: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  new:         { color: "gold",   label: "Новая заявка" },
  confirmed:   { color: "blue",   label: "Подтверждён" },
  in_progress: { color: "cyan",   label: "В работе" },
  ready:       { color: "purple", label: "Готов" },
  delivered:   { color: "green",  label: "Выдан" },
  cancelled:   { color: "red",    label: "Отменён" },
};

// Allowed transitions: current status → [next statuses]
const TRANSITIONS: Record<string, { status: string; label: string; danger?: boolean }[]> = {
  new:         [{ status: "cancelled", label: "Отменить", danger: true }],
  confirmed:   [
    { status: "in_progress", label: "Взять в работу" },
    { status: "cancelled",   label: "Отменить", danger: true },
  ],
  in_progress: [
    { status: "ready",     label: "Готово к выдаче" },
    { status: "cancelled", label: "Отменить", danger: true },
  ],
  ready:       [
    { status: "delivered", label: "Выдан клиенту" },
    { status: "cancelled", label: "Отменить", danger: true },
  ],
};

const formatPrice = (price: number) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency", currency: "RUB", maximumFractionDigits: 0,
  }).format(price);

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "—";
  return new Date(dateStr).toLocaleDateString("ru-RU", {
    day: "2-digit", month: "2-digit", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
};

export const CustomOrderDetail = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const [order, setOrder] = useState<CustomOrder | null>(null);
  const [loading, setLoading] = useState(true);

  // Confirm modal state
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirmForm] = Form.useForm();
  const [confirmLoading, setConfirmLoading] = useState(false);

  // Admin notes edit state
  const [notesOpen, setNotesOpen] = useState(false);
  const [notesForm] = Form.useForm();
  const [notesLoading, setNotesLoading] = useState(false);

  // File upload state
  const [uploadLoading, setUploadLoading] = useState(false);
  const [paymentLoading, setPaymentLoading] = useState(false);
  const [statusLoading, setStatusLoading] = useState(false);

  const fetchOrder = useCallback(async () => {
    setLoading(true);
    try {
      const { data: resp } = await api.get(`/admin/custom-orders/${id}`);
      setOrder(resp.data);
    } catch {
      message.error("Не удалось загрузить заказ");
    } finally {
      setLoading(false);
    }
  }, [id, message]);

  useEffect(() => { fetchOrder(); }, [fetchOrder]);

  // ── Status transitions ──────────────────────────────────────────────────
  const handleStatusChange = async (newStatus: string) => {
    if (!order) return;
    setStatusLoading(true);
    try {
      await api.post(`/admin/orders/${order.id}/status`, { status: newStatus });
      message.success("Статус обновлён");
      fetchOrder();
    } catch {
      message.error("Ошибка изменения статуса");
    } finally {
      setStatusLoading(false);
    }
  };

  // ── Confirm order (set price) ───────────────────────────────────────────
  const handleConfirm = async (values: { totalPrice: number; adminNotes?: string }) => {
    if (!order) return;
    setConfirmLoading(true);
    try {
      await api.post(`/admin/custom-orders/${order.id}/confirm`, values);
      message.success("Заказ подтверждён, цена установлена");
      setConfirmOpen(false);
      confirmForm.resetFields();
      fetchOrder();
    } catch {
      message.error("Ошибка подтверждения заказа");
    } finally {
      setConfirmLoading(false);
    }
  };

  // ── Admin notes update ──────────────────────────────────────────────────
  const handleNotesUpdate = async (values: { adminNotes?: string }) => {
    if (!order) return;
    setNotesLoading(true);
    try {
      await api.put(`/admin/custom-orders/${order.id}`, values);
      message.success("Заметки сохранены");
      setNotesOpen(false);
      fetchOrder();
    } catch {
      message.error("Ошибка сохранения");
    } finally {
      setNotesLoading(false);
    }
  };

  // ── Payment ─────────────────────────────────────────────────────────────
  const handleSendPayment = async () => {
    if (!order) return;
    setPaymentLoading(true);
    try {
      await api.post(`/admin/custom-orders/${order.id}/send-payment`);
      message.success("Ссылка на оплату создана");
      fetchOrder();
    } catch {
      message.error("Ошибка создания ссылки на оплату");
    } finally {
      setPaymentLoading(false);
    }
  };

  const handleMarkPaid = async () => {
    if (!order) return;
    try {
      await api.post(`/admin/custom-orders/${order.id}/mark-paid`);
      message.success("Заказ отмечен как оплаченный");
      fetchOrder();
    } catch {
      message.error("Ошибка");
    }
  };

  // ── File upload ─────────────────────────────────────────────────────────
  const handleFileUpload = async (file: File) => {
    if (!order) return false;
    setUploadLoading(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      await api.post(`/admin/custom-orders/${order.id}/files`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      message.success(`${file.name} загружен`);
      fetchOrder();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        "Ошибка загрузки файла";
      message.error(msg);
    } finally {
      setUploadLoading(false);
    }
    return false; // prevent default antd upload behavior
  };

  const handleFileDelete = async (url: string) => {
    if (!order) return;
    try {
      await api.delete(`/admin/custom-orders/${order.id}/files`, { data: { url } });
      message.success("Файл удалён");
      fetchOrder();
    } catch {
      message.error("Ошибка удаления файла");
    }
  };

  // ── Items table ─────────────────────────────────────────────────────────
  const itemColumns: ColumnsType<OrderItem> = [
    {
      title: "Наименование",
      key: "name",
      render: (_: unknown, r: OrderItem) => (
        <div>
          <div style={{ fontWeight: 500 }}>
            {r.customItemName || `Товар #${r.productId}`}
          </div>
          {r.customItemDescription && (
            <div style={{ fontSize: 12, color: "#888" }}>{r.customItemDescription}</div>
          )}
        </div>
      ),
    },
    {
      title: "Кол-во",
      dataIndex: "quantity",
      key: "qty",
      width: 80,
      align: "center",
    },
    {
      title: "Цена",
      dataIndex: "unitPrice",
      key: "unitPrice",
      width: 110,
      render: (p: number) => formatPrice(p),
    },
    {
      title: "Итого",
      dataIndex: "totalPrice",
      key: "total",
      width: 120,
      render: (p: number) => <strong>{formatPrice(p)}</strong>,
    },
  ];

  if (loading || !order) {
    return <Card loading style={{ minHeight: 400 }} />;
  }

  const transitions = TRANSITIONS[order.status] || [];
  const details = order.customDetails;
  const fileUrls: string[] = details?.fileUrls || [];

  return (
    <>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        {/* Header */}
        <Card size="small">
          <Space wrap>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/custom-orders")}
            >
              Назад
            </Button>
            <span style={{ fontWeight: 600, fontSize: 16, fontFamily: "monospace" }}>
              {order.orderNumber}
            </span>
            <Tag color={STATUS_CONFIG[order.status]?.color || "default"}>
              {STATUS_CONFIG[order.status]?.label || order.status}
            </Tag>
            {order.isPaid ? (
              <Badge status="success" text="Оплачен" />
            ) : (
              <Badge status="warning" text="Не оплачен" />
            )}
          </Space>
        </Card>

        <Row gutter={16}>
          {/* ── Left column ── */}
          <Col xs={24} lg={16}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>

              {/* Items */}
              <Card title="Позиции заказа" size="small">
                {order.items && order.items.length > 0 ? (
                  <Table
                    columns={itemColumns}
                    dataSource={order.items}
                    rowKey="id"
                    pagination={false}
                    size="small"
                    summary={() => (
                      <Table.Summary.Row>
                        <Table.Summary.Cell index={0} colSpan={3}>
                          <strong>Итого</strong>
                        </Table.Summary.Cell>
                        <Table.Summary.Cell index={1}>
                          <strong>{formatPrice(order.totalPrice)}</strong>
                        </Table.Summary.Cell>
                      </Table.Summary.Row>
                    )}
                  />
                ) : (
                  <Text type="secondary">
                    Позиции не добавлены (заявка без конкретных позиций)
                  </Text>
                )}
              </Card>

              {/* Client description */}
              {details?.clientDescription && (
                <Card title="Описание клиента" size="small">
                  <Text>{details.clientDescription}</Text>
                </Card>
              )}

              {/* Admin notes */}
              <Card
                title="Заметки администратора"
                size="small"
                extra={
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => {
                      notesForm.setFieldsValue({ adminNotes: details?.adminNotes });
                      setNotesOpen(true);
                    }}
                  >
                    Редактировать
                  </Button>
                }
              >
                {details?.adminNotes ? (
                  <Text>{details.adminNotes}</Text>
                ) : (
                  <Text type="secondary">Нет заметок</Text>
                )}
              </Card>

              {/* Files */}
              <Card
                title={`Файлы (${fileUrls.length}/5)`}
                size="small"
                extra={
                  <Upload
                    beforeUpload={handleFileUpload}
                    showUploadList={false}
                    accept=".stl,.obj,.3mf,.step,.stp,.zip"
                    disabled={fileUrls.length >= 5 || uploadLoading}
                  >
                    <Button
                      size="small"
                      icon={<UploadOutlined />}
                      loading={uploadLoading}
                      disabled={fileUrls.length >= 5}
                    >
                      Загрузить
                    </Button>
                  </Upload>
                }
              >
                {fileUrls.length === 0 ? (
                  <Text type="secondary">Файлы не прикреплены</Text>
                ) : (
                  <Space direction="vertical" size={6} style={{ width: "100%" }}>
                    {fileUrls.map((url) => {
                      const name = url.split("/").pop() || url;
                      return (
                        <Space key={url} style={{ width: "100%", justifyContent: "space-between" }}>
                          <Link href={url} target="_blank" ellipsis style={{ maxWidth: 350 }}>
                            {name}
                          </Link>
                          <Popconfirm
                            title="Удалить файл?"
                            onConfirm={() => handleFileDelete(url)}
                            okText="Да"
                            cancelText="Нет"
                          >
                            <Button
                              type="text"
                              size="small"
                              danger
                              icon={<DeleteOutlined />}
                            />
                          </Popconfirm>
                        </Space>
                      );
                    })}
                  </Space>
                )}
              </Card>

            </Space>
          </Col>

          {/* ── Right column ── */}
          <Col xs={24} lg={8}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>

              {/* Actions */}
              <Card title="Действия" size="small">
                <Space direction="vertical" style={{ width: "100%" }}>
                  {order.status === "new" && (
                    <Button
                      type="primary"
                      icon={<CheckOutlined />}
                      block
                      onClick={() => {
                        confirmForm.resetFields();
                        setConfirmOpen(true);
                      }}
                    >
                      Подтвердить и установить цену
                    </Button>
                  )}

                  {transitions.map((t) => (
                    t.danger ? (
                      <Popconfirm
                        key={t.status}
                        title={`Перевести в статус "${STATUS_CONFIG[t.status]?.label}"?`}
                        onConfirm={() => handleStatusChange(t.status)}
                        okText="Да"
                        cancelText="Нет"
                        okButtonProps={{ danger: true }}
                      >
                        <Button
                          danger
                          icon={<CloseOutlined />}
                          block
                          loading={statusLoading}
                        >
                          {t.label}
                        </Button>
                      </Popconfirm>
                    ) : t.status !== "new" ? (
                      <Button
                        key={t.status}
                        type="default"
                        icon={<SyncOutlined />}
                        block
                        loading={statusLoading}
                        onClick={() => handleStatusChange(t.status)}
                      >
                        {t.label}
                      </Button>
                    ) : null
                  ))}
                </Space>
              </Card>

              {/* Payment */}
              <Card title="Оплата" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Способ">
                    {order.paymentMethod === "card" ? "Карта онлайн" : "Наличные"}
                  </Descriptions.Item>
                  <Descriptions.Item label="Статус">
                    {order.isPaid ? (
                      <Tag color="green">Оплачен</Tag>
                    ) : (
                      <Tag color="orange">Не оплачен</Tag>
                    )}
                  </Descriptions.Item>
                  {order.totalPrice > 0 && (
                    <Descriptions.Item label="Сумма">
                      <strong>{formatPrice(order.totalPrice)}</strong>
                    </Descriptions.Item>
                  )}
                  {order.paymentLink && (
                    <Descriptions.Item label="Ссылка">
                      <Link href={order.paymentLink} target="_blank" copyable>
                        <LinkOutlined /> Открыть
                      </Link>
                    </Descriptions.Item>
                  )}
                  {order.paymentExpiresAt && (
                    <Descriptions.Item label="Истекает">
                      {formatDate(order.paymentExpiresAt)}
                    </Descriptions.Item>
                  )}
                </Descriptions>
                {!order.isPaid && order.status !== "cancelled" && (
                  <Space style={{ marginTop: 8 }} wrap>
                    {order.paymentMethod === "card" && order.totalPrice > 0 && (
                      <Button
                        size="small"
                        icon={<SendOutlined />}
                        loading={paymentLoading}
                        onClick={handleSendPayment}
                      >
                        {order.paymentLink ? "Обновить ссылку" : "Создать ссылку"}
                      </Button>
                    )}
                    <Popconfirm
                      title="Отметить заказ как оплаченный?"
                      onConfirm={handleMarkPaid}
                      okText="Да"
                      cancelText="Нет"
                    >
                      <Button size="small" icon={<DollarOutlined />}>
                        Отметить оплаченным
                      </Button>
                    </Popconfirm>
                  </Space>
                )}
              </Card>

              {/* Customer */}
              <Card title="Клиент" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Имя">
                    {order.customerName}
                  </Descriptions.Item>
                  <Descriptions.Item label="Телефон">
                    <a href={`tel:${order.customerPhone}`}>{order.customerPhone}</a>
                  </Descriptions.Item>
                  {order.customerEmail && (
                    <Descriptions.Item label="Email">
                      <a href={`mailto:${order.customerEmail}`}>
                        {order.customerEmail}
                      </a>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>

              {/* Delivery */}
              <Card title="Доставка" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Способ">
                    {order.deliveryMethod === "courier"
                      ? "Курьер"
                      : order.deliveryMethod === "pickup_point"
                      ? "Пункт выдачи"
                      : "Самовывоз"}
                  </Descriptions.Item>
                  {order.deliveryAddress && (
                    <Descriptions.Item label="Адрес">
                      {order.deliveryAddress}
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>

              {/* Bitrix */}
              {(details?.bitrixDealId || details?.bitrixStageId) && (
                <Card title="Bitrix24" size="small">
                  <Descriptions column={1} size="small">
                    {details.bitrixDealId && (
                      <Descriptions.Item label="Сделка">
                        <Tag color="geekblue">#{details.bitrixDealId}</Tag>
                      </Descriptions.Item>
                    )}
                    {details.bitrixStageId && (
                      <Descriptions.Item label="Стадия">
                        {details.bitrixStageId}
                      </Descriptions.Item>
                    )}
                  </Descriptions>
                </Card>
              )}

              {/* Dates */}
              <Card title="Информация" size="small">
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="Создан">
                    {formatDate(order.createdAt)}
                  </Descriptions.Item>
                  {order.notes && (
                    <Descriptions.Item label="Примечание">
                      {order.notes}
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>

            </Space>
          </Col>
        </Row>
      </Space>

      {/* ── Confirm modal ── */}
      <Modal
        title="Подтверждение заказа и установка цены"
        open={confirmOpen}
        onCancel={() => setConfirmOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={confirmForm}
          layout="vertical"
          onFinish={handleConfirm}
        >
          <Form.Item
            name="totalPrice"
            label="Итоговая стоимость (₽)"
            rules={[{ required: true, message: "Укажите стоимость" }]}
          >
            <InputNumber
              min={1}
              style={{ width: "100%" }}
              placeholder="Например: 3500"
              addonAfter="₽"
            />
          </Form.Item>
          <Form.Item name="adminNotes" label="Заметка для клиента (необязательно)">
            <TextArea rows={3} placeholder="Комментарий к заказу…" />
          </Form.Item>
          <Divider />
          <Space style={{ width: "100%", justifyContent: "flex-end" }}>
            <Button onClick={() => setConfirmOpen(false)}>Отмена</Button>
            <Button type="primary" htmlType="submit" loading={confirmLoading}>
              Подтвердить
            </Button>
          </Space>
        </Form>
      </Modal>

      {/* ── Admin notes modal ── */}
      <Modal
        title="Заметки администратора"
        open={notesOpen}
        onCancel={() => setNotesOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form
          form={notesForm}
          layout="vertical"
          onFinish={handleNotesUpdate}
        >
          <Form.Item name="adminNotes" label="Заметки">
            <TextArea rows={5} placeholder="Внутренние заметки по заказу…" />
          </Form.Item>
          <Space style={{ width: "100%", justifyContent: "flex-end" }}>
            <Button onClick={() => setNotesOpen(false)}>Отмена</Button>
            <Button type="primary" htmlType="submit" loading={notesLoading}>
              Сохранить
            </Button>
          </Space>
        </Form>
      </Modal>
    </>
  );
};

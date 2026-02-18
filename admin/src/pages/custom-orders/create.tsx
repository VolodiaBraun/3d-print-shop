import { useState } from "react";
import {
  Card, Form, Input, Select, Button, Space, InputNumber,
  Table, App, Divider, Row, Col,
} from "antd";
import {
  ArrowLeftOutlined, PlusOutlined, DeleteOutlined, SaveOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import api from "../../lib/api";

const { TextArea } = Input;

interface ItemRow {
  key: number;
  name: string;
  description?: string;
  quantity: number;
  unitPrice: number;
}

let itemKeyCounter = 1;

export const CustomOrderCreate = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<ItemRow[]>([]);
  const [deliveryMethod, setDeliveryMethod] = useState("pickup");

  const addItem = () => {
    setItems((prev) => [
      ...prev,
      { key: itemKeyCounter++, name: "", quantity: 1, unitPrice: 0 },
    ]);
  };

  const removeItem = (key: number) =>
    setItems((prev) => prev.filter((i) => i.key !== key));

  const updateItem = (key: number, field: keyof ItemRow, value: unknown) =>
    setItems((prev) =>
      prev.map((i) => (i.key === key ? { ...i, [field]: value } : i))
    );

  const subtotal = items.reduce((sum, i) => sum + i.quantity * i.unitPrice, 0);

  const handleSubmit = async (values: Record<string, unknown>) => {
    // Validate items if any
    for (const item of items) {
      if (!item.name.trim()) {
        message.error("Заполните наименование для всех позиций");
        return;
      }
    }

    setLoading(true);
    try {
      const payload = {
        customerName: values.customerName,
        customerPhone: values.customerPhone,
        customerEmail: values.customerEmail || undefined,
        clientDescription: values.clientDescription || undefined,
        adminNotes: values.adminNotes || undefined,
        paymentMethod: values.paymentMethod,
        deliveryMethod: values.deliveryMethod,
        deliveryAddress: values.deliveryAddress || undefined,
        notes: values.notes || undefined,
        bitrixDealId: values.bitrixDealId || undefined,
        items: items.map((i) => ({
          name: i.name,
          description: i.description || undefined,
          quantity: i.quantity,
          unitPrice: i.unitPrice,
        })),
      };

      const { data: resp } = await api.post("/admin/custom-orders", payload);
      message.success("Заказ создан");
      navigate(`/custom-orders/${resp.data?.id}`);
    } catch {
      message.error("Ошибка создания заказа");
    } finally {
      setLoading(false);
    }
  };

  const itemColumns = [
    {
      title: "Наименование *",
      key: "name",
      render: (_: unknown, r: ItemRow) => (
        <Input
          value={r.name}
          placeholder="Печать детали, пост-обработка…"
          onChange={(e) => updateItem(r.key, "name", e.target.value)}
        />
      ),
    },
    {
      title: "Описание",
      key: "description",
      render: (_: unknown, r: ItemRow) => (
        <Input
          value={r.description}
          placeholder="Необязательно"
          onChange={(e) => updateItem(r.key, "description", e.target.value)}
        />
      ),
    },
    {
      title: "Кол-во",
      key: "qty",
      width: 90,
      render: (_: unknown, r: ItemRow) => (
        <InputNumber
          min={1}
          value={r.quantity}
          style={{ width: "100%" }}
          onChange={(v) => updateItem(r.key, "quantity", v ?? 1)}
        />
      ),
    },
    {
      title: "Цена (₽)",
      key: "price",
      width: 120,
      render: (_: unknown, r: ItemRow) => (
        <InputNumber
          min={0}
          value={r.unitPrice}
          style={{ width: "100%" }}
          onChange={(v) => updateItem(r.key, "unitPrice", v ?? 0)}
        />
      ),
    },
    {
      title: "Итого",
      key: "total",
      width: 110,
      render: (_: unknown, r: ItemRow) => (
        <span style={{ fontWeight: 500 }}>
          {new Intl.NumberFormat("ru-RU", {
            style: "currency",
            currency: "RUB",
            maximumFractionDigits: 0,
          }).format(r.quantity * r.unitPrice)}
        </span>
      ),
    },
    {
      title: "",
      key: "del",
      width: 44,
      render: (_: unknown, r: ItemRow) => (
        <Button
          type="text"
          danger
          icon={<DeleteOutlined />}
          onClick={() => removeItem(r.key)}
        />
      ),
    },
  ];

  return (
    <Form form={form} layout="vertical" onFinish={handleSubmit}>
      <Space direction="vertical" size={16} style={{ width: "100%" }}>

        {/* Header */}
        <Card size="small">
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/custom-orders")}
            >
              Назад
            </Button>
            <span style={{ fontWeight: 600, fontSize: 16 }}>
              Новый индивидуальный заказ
            </span>
          </Space>
        </Card>

        <Row gutter={16}>
          {/* Left */}
          <Col xs={24} lg={16}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>

              {/* Customer */}
              <Card title="Клиент" size="small">
                <Row gutter={16}>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="customerName"
                      label="Имя"
                      rules={[{ required: true, message: "Введите имя" }]}
                    >
                      <Input placeholder="Иван Иванов" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="customerPhone"
                      label="Телефон"
                      rules={[{ required: true, message: "Введите телефон" }]}
                    >
                      <Input placeholder="+7 900 000-00-00" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item name="customerEmail" label="Email">
                      <Input placeholder="mail@example.com" />
                    </Form.Item>
                  </Col>
                </Row>
              </Card>

              {/* Description */}
              <Card title="Описание заказа" size="small">
                <Form.Item name="clientDescription" label="Что хочет клиент">
                  <TextArea
                    rows={3}
                    placeholder="Описание задачи, требования к материалу, цвету…"
                  />
                </Form.Item>
                <Form.Item name="adminNotes" label="Внутренние заметки">
                  <TextArea rows={2} placeholder="Заметки для себя" />
                </Form.Item>
              </Card>

              {/* Items */}
              <Card
                title="Позиции заказа"
                size="small"
                extra={
                  <Button
                    size="small"
                    icon={<PlusOutlined />}
                    onClick={addItem}
                  >
                    Добавить
                  </Button>
                }
              >
                {items.length > 0 ? (
                  <>
                    <Table
                      columns={itemColumns}
                      dataSource={items}
                      rowKey="key"
                      pagination={false}
                      size="small"
                      scroll={{ x: 600 }}
                    />
                    {subtotal > 0 && (
                      <div style={{ textAlign: "right", marginTop: 8, fontWeight: 600 }}>
                        Итого:{" "}
                        {new Intl.NumberFormat("ru-RU", {
                          style: "currency",
                          currency: "RUB",
                          maximumFractionDigits: 0,
                        }).format(subtotal)}
                      </div>
                    )}
                  </>
                ) : (
                  <div style={{ color: "#999", textAlign: "center", padding: "12px 0" }}>
                    Позиции не добавлены — цена будет установлена при подтверждении заявки
                  </div>
                )}
              </Card>

            </Space>
          </Col>

          {/* Right */}
          <Col xs={24} lg={8}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>

              <Card title="Оплата и доставка" size="small">
                <Form.Item
                  name="paymentMethod"
                  label="Способ оплаты"
                  initialValue="card"
                  rules={[{ required: true }]}
                >
                  <Select
                    options={[
                      { value: "card", label: "Карта онлайн" },
                      { value: "cash", label: "Наличные / перевод" },
                    ]}
                  />
                </Form.Item>

                <Form.Item
                  name="deliveryMethod"
                  label="Способ получения"
                  initialValue="pickup"
                  rules={[{ required: true }]}
                >
                  <Select
                    options={[
                      { value: "pickup", label: "Самовывоз" },
                      { value: "courier", label: "Курьер" },
                      { value: "pickup_point", label: "Пункт выдачи" },
                    ]}
                    onChange={setDeliveryMethod}
                  />
                </Form.Item>

                {(deliveryMethod === "courier" ||
                  deliveryMethod === "pickup_point") && (
                  <Form.Item name="deliveryAddress" label="Адрес">
                    <Input placeholder="Улица, дом, квартира" />
                  </Form.Item>
                )}

                <Form.Item name="notes" label="Примечание к заказу">
                  <Input placeholder="Необязательно" />
                </Form.Item>
              </Card>

              <Card title="Bitrix24" size="small">
                <Form.Item name="bitrixDealId" label="ID сделки">
                  <Input placeholder="Оставьте пустым для авто-создания" />
                </Form.Item>
              </Card>

            </Space>
          </Col>
        </Row>

        {/* Footer */}
        <Card size="small">
          <Divider style={{ margin: "0 0 12px" }} />
          <Space style={{ justifyContent: "flex-end", width: "100%", display: "flex" }}>
            <Button onClick={() => navigate("/custom-orders")}>Отмена</Button>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SaveOutlined />}
              loading={loading}
            >
              Создать заказ
            </Button>
          </Space>
        </Card>

      </Space>
    </Form>
  );
};

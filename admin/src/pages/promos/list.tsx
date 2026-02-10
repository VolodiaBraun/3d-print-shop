import { useState, useEffect, useCallback } from "react";
import {
  Table,
  Button,
  Space,
  Tag,
  Card,
  App,
  Popconfirm,
  Switch,
} from "antd";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  GiftOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useNavigate } from "react-router-dom";
import api from "../../lib/api";

interface PromoCode {
  id: number;
  code: string;
  description?: string;
  discountType: string;
  discountValue: number;
  minOrderAmount: number;
  maxUses: number;
  usedCount: number;
  isActive: boolean;
  startsAt?: string;
  expiresAt?: string;
  createdAt: string;
}

export const PromoList = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [promos, setPromos] = useState<PromoCode[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPromos = useCallback(async () => {
    setLoading(true);
    try {
      const { data: resp } = await api.get("/admin/promo-codes");
      setPromos(resp.data || []);
    } catch {
      message.error("Не удалось загрузить промокоды");
    } finally {
      setLoading(false);
    }
  }, [message]);

  useEffect(() => {
    fetchPromos();
  }, [fetchPromos]);

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/promo-codes/${id}`);
      message.success("Промокод удалён");
      fetchPromos();
    } catch {
      message.error("Не удалось удалить промокод");
    }
  };

  const handleToggleActive = async (id: number, active: boolean) => {
    try {
      await api.put(`/admin/promo-codes/${id}`, { isActive: active });
      message.success(active ? "Промокод активирован" : "Промокод деактивирован");
      fetchPromos();
    } catch {
      message.error("Ошибка обновления статуса");
    }
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat("ru-RU", {
      style: "currency",
      currency: "RUB",
      maximumFractionDigits: 0,
    }).format(price);

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    });
  };

  const columns: ColumnsType<PromoCode> = [
    {
      title: "Код",
      dataIndex: "code",
      key: "code",
      width: 160,
      render: (code: string) => (
        <span style={{ fontWeight: 600, fontFamily: "monospace", fontSize: 14 }}>
          {code}
        </span>
      ),
    },
    {
      title: "Скидка",
      key: "discount",
      width: 130,
      render: (_: unknown, record: PromoCode) => (
        <Tag color="blue" style={{ fontSize: 13 }}>
          {record.discountType === "percent"
            ? `${record.discountValue}%`
            : formatPrice(record.discountValue)}
        </Tag>
      ),
    },
    {
      title: "Мин. сумма",
      dataIndex: "minOrderAmount",
      key: "minOrderAmount",
      width: 120,
      render: (val: number) =>
        val > 0 ? (
          <span style={{ color: "#666" }}>{formatPrice(val)}</span>
        ) : (
          <span style={{ color: "#ccc" }}>—</span>
        ),
    },
    {
      title: "Использований",
      key: "usage",
      width: 140,
      align: "center",
      render: (_: unknown, record: PromoCode) => (
        <span>
          <strong>{record.usedCount}</strong>
          {record.maxUses > 0 && (
            <span style={{ color: "#999" }}> / {record.maxUses}</span>
          )}
        </span>
      ),
    },
    {
      title: "Период",
      key: "period",
      width: 180,
      render: (_: unknown, record: PromoCode) => (
        <div style={{ fontSize: 12, color: "#666" }}>
          {record.startsAt || record.expiresAt ? (
            <>
              {formatDate(record.startsAt)} — {formatDate(record.expiresAt)}
            </>
          ) : (
            <span style={{ color: "#ccc" }}>Бессрочный</span>
          )}
        </div>
      ),
    },
    {
      title: "Активен",
      dataIndex: "isActive",
      key: "isActive",
      width: 90,
      align: "center",
      render: (active: boolean, record: PromoCode) => (
        <Switch
          checked={active}
          size="small"
          onChange={(checked) => handleToggleActive(record.id, checked)}
        />
      ),
    },
    {
      title: "Действия",
      key: "actions",
      width: 100,
      render: (_: unknown, record: PromoCode) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => navigate(`/promos/${record.id}/edit`)}
          />
          <Popconfirm
            title="Удалить промокод?"
            onConfirm={() => handleDelete(record.id)}
            okText="Удалить"
            cancelText="Отмена"
            okButtonProps={{ danger: true }}
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <GiftOutlined />
          <span>Промокоды</span>
        </Space>
      }
      extra={
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate("/promos/create")}
        >
          Создать промокод
        </Button>
      }
    >
      <Table
        columns={columns}
        dataSource={promos}
        rowKey="id"
        loading={loading}
        pagination={false}
        scroll={{ x: 800 }}
        size="middle"
      />
    </Card>
  );
};

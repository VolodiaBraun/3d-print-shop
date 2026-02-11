import { useState, useEffect, useCallback } from "react";
import { Table, Tag, Select, Card, Space, Button, App, Rate, Popconfirm } from "antd";
import {
  MessageOutlined,
  CheckOutlined,
  CloseOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import api from "../../lib/api";

interface ReviewUser {
  id: number;
  firstName?: string;
  lastName?: string;
  email?: string;
}

interface ReviewProduct {
  id: number;
  name: string;
}

interface Review {
  id: number;
  userId: number;
  productId: number;
  orderId: number;
  rating: number;
  comment: string;
  status: string;
  user?: ReviewUser;
  product?: ReviewProduct;
  createdAt: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  pending: { color: "gold", label: "На модерации" },
  approved: { color: "green", label: "Одобрен" },
  rejected: { color: "red", label: "Отклонён" },
};

const STATUS_OPTIONS = [
  { value: "", label: "Все статусы" },
  { value: "pending", label: "На модерации" },
  { value: "approved", label: "Одобрен" },
  { value: "rejected", label: "Отклонён" },
];

export const ReviewList = () => {
  const { message } = App.useApp();
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState("");

  const fetchReviews = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = {
        page,
        limit: pageSize,
      };
      if (statusFilter) params.status = statusFilter;

      const { data: resp } = await api.get("/admin/reviews", { params });
      setReviews(resp.data || []);
      setTotal(resp.meta?.total || 0);
    } catch {
      message.error("Не удалось загрузить отзывы");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, statusFilter, message]);

  useEffect(() => {
    fetchReviews();
  }, [fetchReviews]);

  const handleApprove = async (id: number) => {
    try {
      await api.put(`/admin/reviews/${id}/approve`);
      message.success("Отзыв одобрен");
      fetchReviews();
    } catch {
      message.error("Ошибка при одобрении");
    }
  };

  const handleReject = async (id: number) => {
    try {
      await api.put(`/admin/reviews/${id}/reject`);
      message.success("Отзыв отклонён");
      fetchReviews();
    } catch {
      message.error("Ошибка при отклонении");
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/reviews/${id}`);
      message.success("Отзыв удалён");
      fetchReviews();
    } catch {
      message.error("Ошибка при удалении");
    }
  };

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getUserName = (user?: ReviewUser) => {
    if (!user) return "—";
    const parts = [user.firstName, user.lastName].filter(Boolean);
    return parts.length > 0 ? parts.join(" ") : user.email || `#${user.id}`;
  };

  const columns: ColumnsType<Review> = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      width: 60,
    },
    {
      title: "Товар",
      key: "product",
      width: 200,
      render: (_: unknown, record: Review) => (
        <span style={{ fontWeight: 500 }}>
          {record.product?.name || `Товар #${record.productId}`}
        </span>
      ),
    },
    {
      title: "Пользователь",
      key: "user",
      width: 160,
      render: (_: unknown, record: Review) => getUserName(record.user),
    },
    {
      title: "Рейтинг",
      dataIndex: "rating",
      key: "rating",
      width: 140,
      render: (rating: number) => (
        <Rate disabled value={rating} style={{ fontSize: 14 }} />
      ),
    },
    {
      title: "Комментарий",
      dataIndex: "comment",
      key: "comment",
      ellipsis: true,
      render: (text: string) => text || <span style={{ color: "#999" }}>—</span>,
    },
    {
      title: "Статус",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status: string) => {
        const cfg = STATUS_CONFIG[status] || { color: "default", label: status };
        return <Tag color={cfg.color}>{cfg.label}</Tag>;
      },
    },
    {
      title: "Дата",
      dataIndex: "createdAt",
      key: "createdAt",
      width: 150,
      render: (date: string) => (
        <span style={{ fontSize: 13, color: "#666" }}>{formatDate(date)}</span>
      ),
    },
    {
      title: "Действия",
      key: "actions",
      width: 130,
      render: (_: unknown, record: Review) => (
        <Space size="small">
          {record.status !== "approved" && (
            <Button
              type="text"
              size="small"
              icon={<CheckOutlined />}
              style={{ color: "#52c41a" }}
              title="Одобрить"
              onClick={() => handleApprove(record.id)}
            />
          )}
          {record.status !== "rejected" && (
            <Button
              type="text"
              size="small"
              icon={<CloseOutlined />}
              style={{ color: "#faad14" }}
              title="Отклонить"
              onClick={() => handleReject(record.id)}
            />
          )}
          <Popconfirm
            title="Удалить отзыв?"
            onConfirm={() => handleDelete(record.id)}
            okText="Да"
            cancelText="Нет"
          >
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              danger
              title="Удалить"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const handleTableChange = (pagination: TablePaginationConfig) => {
    setPage(pagination.current || 1);
    setPageSize(pagination.pageSize || 20);
  };

  return (
    <Card
      title={
        <Space>
          <MessageOutlined />
          <span>Отзывы</span>
        </Space>
      }
    >
      <Space style={{ marginBottom: 16 }} wrap>
        <Select
          style={{ width: 200 }}
          options={STATUS_OPTIONS}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val);
            setPage(1);
          }}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={reviews}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `Всего: ${t}`,
        }}
        onChange={handleTableChange}
        rowClassName={(record) =>
          record.status === "pending" ? "review-pending-row" : ""
        }
        scroll={{ x: 1000 }}
        size="middle"
      />

      <style>{`
        .review-pending-row {
          background-color: #fffbe6 !important;
        }
        .review-pending-row:hover > td {
          background-color: #fff7cc !important;
        }
      `}</style>
    </Card>
  );
};

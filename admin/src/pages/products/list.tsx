import { useState, useEffect, useCallback } from "react";
import {
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  Card,
  Image,
  App,
  Popconfirm,
} from "antd";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  ShoppingOutlined,
} from "@ant-design/icons";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useNavigate } from "react-router-dom";
import api from "../../lib/api";

interface ProductImage {
  id: number;
  url: string;
  urlThumbnail?: string;
  isMain: boolean;
}

interface Category {
  id: number;
  name: string;
  slug: string;
}

interface Product {
  id: number;
  name: string;
  slug: string;
  price: number;
  oldPrice?: number;
  stockQuantity: number;
  isActive: boolean;
  isFeatured: boolean;
  categoryId?: number;
  category?: Category;
  images?: ProductImage[];
  material?: string;
  createdAt: string;
}

interface CategoryOption {
  id: number;
  name: string;
  children?: CategoryOption[];
}

export const ProductList = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string>("");
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);
  const [bulkLoading, setBulkLoading] = useState(false);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = {
        page,
        limit: pageSize,
      };
      if (search) params.search = search;
      if (categoryFilter) params.category = categoryFilter;

      const { data: resp } = await api.get("/admin/products", { params });
      setProducts(resp.data || []);
      setTotal(resp.meta?.total || 0);
    } catch {
      message.error("Не удалось загрузить товары");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, search, categoryFilter, message]);

  const fetchCategories = useCallback(async () => {
    try {
      const { data: resp } = await api.get("/categories");
      setCategories(resp.data || []);
    } catch {
      // silent
    }
  }, []);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/products/${id}`);
      message.success("Товар удалён");
      fetchProducts();
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message || "Не удалось удалить товар"
      );
    }
  };

  const handleBulkDeactivate = async () => {
    if (selectedRowKeys.length === 0) return;
    setBulkLoading(true);
    try {
      await Promise.all(
        selectedRowKeys.map((id) =>
          api.put(`/admin/products/${id}`, { isActive: false })
        )
      );
      message.success(`Деактивировано: ${selectedRowKeys.length} товаров`);
      setSelectedRowKeys([]);
      fetchProducts();
    } catch {
      message.error("Ошибка при деактивации");
    } finally {
      setBulkLoading(false);
    }
  };

  const flattenCategories = (
    cats: CategoryOption[],
    depth = 0
  ): { value: string; label: string }[] => {
    const result: { value: string; label: string }[] = [];
    for (const cat of cats) {
      const prefix = depth > 0 ? "\u2014 ".repeat(depth) : "";
      result.push({
        value: cat.name.toLowerCase(),
        label: `${prefix}${cat.name}`,
      });
      if (cat.children?.length) {
        result.push(...flattenCategories(cat.children, depth + 1));
      }
    }
    return result;
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat("ru-RU", {
      style: "currency",
      currency: "RUB",
      maximumFractionDigits: 0,
    }).format(price);

  const columns: ColumnsType<Product> = [
    {
      title: "Фото",
      dataIndex: "images",
      key: "image",
      width: 70,
      render: (images: ProductImage[]) => {
        const main = images?.find((i) => i.isMain) || images?.[0];
        return main ? (
          <Image
            src={main.urlThumbnail || main.url}
            alt=""
            width={48}
            height={48}
            style={{ objectFit: "cover", borderRadius: 4 }}
            preview={false}
          />
        ) : (
          <div
            style={{
              width: 48,
              height: 48,
              background: "#f0f0f0",
              borderRadius: 4,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <ShoppingOutlined style={{ color: "#ccc" }} />
          </div>
        );
      },
    },
    {
      title: "Название",
      dataIndex: "name",
      key: "name",
      ellipsis: true,
      render: (name: string, record: Product) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          <div style={{ fontSize: 12, color: "#999" }}>/{record.slug}</div>
        </div>
      ),
    },
    {
      title: "Цена",
      dataIndex: "price",
      key: "price",
      width: 140,
      render: (price: number, record: Product) => (
        <div>
          <div style={{ fontWeight: 500 }}>{formatPrice(price)}</div>
          {record.oldPrice && (
            <div
              style={{
                fontSize: 12,
                color: "#999",
                textDecoration: "line-through",
              }}
            >
              {formatPrice(record.oldPrice)}
            </div>
          )}
        </div>
      ),
    },
    {
      title: "Остаток",
      dataIndex: "stockQuantity",
      key: "stock",
      width: 100,
      render: (qty: number) => (
        <Tag color={qty > 5 ? "green" : qty > 0 ? "orange" : "red"}>
          {qty} шт
        </Tag>
      ),
    },
    {
      title: "Категория",
      dataIndex: "category",
      key: "category",
      width: 150,
      render: (cat?: Category) =>
        cat ? <Tag>{cat.name}</Tag> : <span style={{ color: "#ccc" }}>&mdash;</span>,
    },
    {
      title: "Статус",
      dataIndex: "isActive",
      key: "status",
      width: 110,
      render: (active: boolean, record: Product) => (
        <Space direction="vertical" size={0}>
          <Tag color={active ? "green" : "red"}>
            {active ? "Активен" : "Скрыт"}
          </Tag>
          {record.isFeatured && <Tag color="blue">Featured</Tag>}
        </Space>
      ),
    },
    {
      title: "Действия",
      key: "actions",
      width: 100,
      render: (_: unknown, record: Product) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => navigate(`/products/${record.id}/edit`)}
          />
          <Popconfirm
            title="Удалить товар?"
            description="Товар будет деактивирован"
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

  const handleTableChange = (pagination: TablePaginationConfig) => {
    setPage(pagination.current || 1);
    setPageSize(pagination.pageSize || 20);
  };

  return (
    <Card
      title={
        <Space>
          <ShoppingOutlined />
          <span>Товары</span>
        </Space>
      }
      extra={
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate("/products/create")}
        >
          Добавить товар
        </Button>
      }
    >
      {/* Filters */}
      <Space style={{ marginBottom: 16 }} wrap>
        <Input
          placeholder="Поиск..."
          prefix={<SearchOutlined />}
          allowClear
          style={{ width: 250 }}
          onPressEnter={(e) => {
            setSearch((e.target as HTMLInputElement).value);
            setPage(1);
          }}
          onChange={(e) => {
            if (!e.target.value) {
              setSearch("");
              setPage(1);
            }
          }}
        />
        <Select
          placeholder="Категория"
          allowClear
          style={{ width: 200 }}
          options={flattenCategories(categories)}
          value={categoryFilter || undefined}
          onChange={(val) => {
            setCategoryFilter(val || "");
            setPage(1);
          }}
        />
        {selectedRowKeys.length > 0 && (
          <Popconfirm
            title={`Деактивировать ${selectedRowKeys.length} товаров?`}
            onConfirm={handleBulkDeactivate}
            okText="Да"
            cancelText="Нет"
          >
            <Button danger loading={bulkLoading}>
              Деактивировать ({selectedRowKeys.length})
            </Button>
          </Popconfirm>
        )}
      </Space>

      <Table
        columns={columns}
        dataSource={products}
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
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys as number[]),
        }}
        scroll={{ x: 800 }}
        size="middle"
      />
    </Card>
  );
};

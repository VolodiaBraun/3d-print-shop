import { useState, useEffect, useCallback } from "react";
import {
  Tree,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Space,
  Popconfirm,
  Tag,
  Spin,
  Typography,
  Card,
  Empty,
  Switch,
  App,
} from "antd";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FolderOutlined,
  FolderOpenOutlined,
} from "@ant-design/icons";
import type { TreeDataNode, TreeProps } from "antd";
import api from "../../lib/api";

const { Text } = Typography;

interface Category {
  id: number;
  name: string;
  slug: string;
  description: string | null;
  parentId: number | null;
  displayOrder: number;
  imageUrl: string | null;
  isActive: boolean;
  children: Category[];
}

type DropInfo = Parameters<NonNullable<TreeProps["onDrop"]>>[0];

export const CategoryList = () => {
  const { message } = App.useApp();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  const fetchCategories = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await api.get("/categories");
      setCategories(data.data);
    } catch {
      message.error("Не удалось загрузить категории");
    } finally {
      setLoading(false);
    }
  }, [message]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  const openCreate = (parentId?: number) => {
    setEditing(null);
    form.resetFields();
    if (parentId) {
      form.setFieldValue("parentId", parentId);
    }
    setModalOpen(true);
  };

  const openEdit = (cat: Category) => {
    setEditing(cat);
    form.setFieldsValue({
      name: cat.name,
      slug: cat.slug,
      description: cat.description || "",
      parentId: cat.parentId || undefined,
      isActive: cat.isActive,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);

      // Clean up empty strings
      if (!values.slug) values.slug = undefined;
      if (!values.description) values.description = undefined;
      if (!values.parentId) values.parentId = undefined;

      if (editing) {
        await api.put(`/admin/categories/${editing.id}`, values);
        message.success("Категория обновлена");
      } else {
        await api.post("/admin/categories", values);
        message.success("Категория создана");
      }

      setModalOpen(false);
      form.resetFields();
      fetchCategories();
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
        errorFields?: unknown;
      };
      if (axiosErr.response?.data?.error?.message) {
        message.error(axiosErr.response.data.error.message);
      }
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/categories/${id}`);
      message.success("Категория удалена");
      fetchCategories();
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message ||
          "Не удалось удалить категорию"
      );
    }
  };

  const handleDrop = async (info: DropInfo) => {
    const dragKey = info.dragNode.key as number;
    const dropKey = info.node.key as number;
    const dropToGap = info.dropToGap;

    let newParentId: number | null;

    if (dropToGap) {
      // Dropped between items — same parent as drop target
      const dropNodeData = findCategory(categories, dropKey);
      newParentId = dropNodeData?.parentId ?? null;
    } else {
      // Dropped onto item — becomes a child
      newParentId = dropKey;
    }

    // Calculate display order based on drop position
    const displayOrder = dropToGap
      ? Math.max(0, info.dropPosition)
      : 0;

    try {
      await api.put(`/admin/categories/${dragKey}`, {
        parentId: newParentId === null ? 0 : newParentId,
        displayOrder,
      });
      message.success("Порядок обновлён");
      fetchCategories();
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message || "Ошибка при перемещении"
      );
    }
  };

  // Helper: find category in tree
  const findCategory = (
    cats: Category[],
    id: number
  ): Category | undefined => {
    for (const cat of cats) {
      if (cat.id === id) return cat;
      if (cat.children?.length) {
        const found = findCategory(cat.children, id);
        if (found) return found;
      }
    }
    return undefined;
  };

  // Convert to Ant Design Tree format
  const toTreeData = (cats: Category[]): TreeDataNode[] => {
    return cats.map((cat) => ({
      key: cat.id,
      title: (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            width: "100%",
            paddingRight: 8,
          }}
        >
          <Space size="small">
            <Text strong>{cat.name}</Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              /{cat.slug}
            </Text>
            {!cat.isActive && <Tag color="red">Скрыта</Tag>}
          </Space>
          <Space size="small">
            <Button
              type="text"
              size="small"
              icon={<PlusOutlined />}
              title="Добавить подкатегорию"
              onClick={(e) => {
                e.stopPropagation();
                openCreate(cat.id);
              }}
            />
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              title="Редактировать"
              onClick={(e) => {
                e.stopPropagation();
                openEdit(cat);
              }}
            />
            <Popconfirm
              title="Удалить категорию?"
              description="Категория с подкатегориями или товарами не может быть удалена"
              onConfirm={() => handleDelete(cat.id)}
              okText="Удалить"
              cancelText="Отмена"
              okButtonProps={{ danger: true }}
            >
              <Button
                type="text"
                size="small"
                danger
                icon={<DeleteOutlined />}
                title="Удалить"
                onClick={(e) => e.stopPropagation()}
              />
            </Popconfirm>
          </Space>
        </div>
      ),
      icon: cat.children?.length ? <FolderOpenOutlined /> : <FolderOutlined />,
      children: cat.children?.length ? toTreeData(cat.children) : undefined,
    }));
  };

  // Flatten categories for parent selector in form
  const flattenForSelect = (
    cats: Category[],
    depth = 0
  ): { value: number; label: string }[] => {
    const result: { value: number; label: string }[] = [];
    for (const cat of cats) {
      const prefix = depth > 0 ? "\u2014 ".repeat(depth) : "";
      result.push({ value: cat.id, label: `${prefix}${cat.name}` });
      if (cat.children?.length) {
        result.push(...flattenForSelect(cat.children, depth + 1));
      }
    }
    return result;
  };

  const parentOptions = flattenForSelect(categories).filter(
    (o) => o.value !== editing?.id
  );

  return (
    <>
      <Card
        title={
          <Space>
            <FolderOutlined />
            <span>Категории</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => openCreate()}
          >
            Добавить категорию
          </Button>
        }
      >
        {loading ? (
          <div style={{ textAlign: "center", padding: 40 }}>
            <Spin size="large" />
          </div>
        ) : categories.length === 0 ? (
          <Empty
            description="Нет категорий"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" onClick={() => openCreate()}>
              Создать первую категорию
            </Button>
          </Empty>
        ) : (
          <Tree
            treeData={toTreeData(categories)}
            draggable
            blockNode
            showIcon
            defaultExpandAll
            onDrop={handleDrop}
            style={{ fontSize: 14 }}
          />
        )}
      </Card>

      <Modal
        title={editing ? "Редактировать категорию" : "Новая категория"}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false);
          form.resetFields();
        }}
        confirmLoading={saving}
        okText={editing ? "Сохранить" : "Создать"}
        cancelText="Отмена"
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          style={{ marginTop: 16 }}
          initialValues={{ isActive: true }}
        >
          <Form.Item
            name="name"
            label="Название"
            rules={[
              { required: true, message: "Введите название" },
              { max: 255, message: "Максимум 255 символов" },
            ]}
          >
            <Input placeholder="Например: Фигурки" />
          </Form.Item>

          <Form.Item
            name="slug"
            label="Slug (URL)"
            help="Оставьте пустым для автогенерации из названия"
          >
            <Input placeholder="figurki" />
          </Form.Item>

          <Form.Item name="description" label="Описание">
            <Input.TextArea rows={3} placeholder="Описание категории" />
          </Form.Item>

          <Form.Item name="parentId" label="Родительская категория">
            <Select
              allowClear
              placeholder="Корневая категория (без родителя)"
              options={parentOptions}
            />
          </Form.Item>

          {editing && (
            <Form.Item
              name="isActive"
              label="Активна"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </>
  );
};

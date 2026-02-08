import { useState, useEffect, useCallback } from "react";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Button,
  Card,
  Row,
  Col,
  Upload,
  Image,
  Space,
  Switch,
  Spin,
  App,
  Typography,
} from "antd";
import {
  SaveOutlined,
  ArrowLeftOutlined,
  PlusOutlined,
  DeleteOutlined,
  StarOutlined,
  StarFilled,
} from "@ant-design/icons";
import type { UploadFile } from "antd";
import { useNavigate, useParams } from "react-router-dom";
import api from "../../lib/api";

const { TextArea } = Input;
const { Title } = Typography;

interface ProductImage {
  id: number;
  url: string;
  urlThumbnail?: string;
  urlMedium?: string;
  isMain: boolean;
  displayOrder: number;
}

interface CategoryOption {
  id: number;
  name: string;
  children?: CategoryOption[];
}

const MATERIALS = [
  { value: "PLA", label: "PLA" },
  { value: "PETG", label: "PETG" },
  { value: "ABS", label: "ABS" },
  { value: "TPU", label: "TPU" },
  { value: "Resin", label: "Смола (Resin)" },
  { value: "Nylon", label: "Нейлон (Nylon)" },
];

export const ProductForm = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);

  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [images, setImages] = useState<ProductImage[]>([]);
  const [uploading, setUploading] = useState(false);

  const fetchCategories = useCallback(async () => {
    try {
      const { data: resp } = await api.get("/categories");
      setCategories(resp.data || []);
    } catch {
      // silent
    }
  }, []);

  const fetchProduct = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const { data: resp } = await api.get(`/admin/products/${id}`);
      const p = resp.data;
      form.setFieldsValue({
        name: p.name,
        slug: p.slug,
        description: p.description || "",
        shortDescription: p.shortDescription || "",
        price: p.price,
        oldPrice: p.oldPrice,
        stockQuantity: p.stockQuantity,
        sku: p.sku || "",
        weight: p.weight,
        dimensionLength: p.dimensions?.length,
        dimensionWidth: p.dimensions?.width,
        dimensionHeight: p.dimensions?.height,
        material: p.material,
        printTime: p.printTime,
        categoryId: p.categoryId,
        isActive: p.isActive,
        isFeatured: p.isFeatured,
      });
      setImages(p.images || []);
    } catch {
      message.error("Не удалось загрузить товар");
      navigate("/products");
    } finally {
      setLoading(false);
    }
  }, [id, form, message, navigate]);

  useEffect(() => {
    fetchCategories();
  }, [fetchCategories]);

  useEffect(() => {
    fetchProduct();
  }, [fetchProduct]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);

      const body: Record<string, unknown> = {
        name: values.name,
        price: values.price,
        stockQuantity: values.stockQuantity ?? 0,
      };

      // Optional fields
      if (values.slug) body.slug = values.slug;
      if (values.description) body.description = values.description;
      if (values.shortDescription)
        body.shortDescription = values.shortDescription;
      if (values.oldPrice !== undefined && values.oldPrice !== null)
        body.oldPrice = values.oldPrice;
      if (values.sku) body.sku = values.sku;
      if (values.weight !== undefined && values.weight !== null)
        body.weight = values.weight;
      if (values.material) body.material = values.material;
      if (values.printTime !== undefined && values.printTime !== null)
        body.printTime = values.printTime;
      if (values.categoryId) body.categoryId = values.categoryId;
      if (values.isFeatured !== undefined) body.isFeatured = values.isFeatured;
      if (isEdit && values.isActive !== undefined)
        body.isActive = values.isActive;

      // Dimensions
      if (
        values.dimensionLength ||
        values.dimensionWidth ||
        values.dimensionHeight
      ) {
        body.dimensions = {
          length: values.dimensionLength || 0,
          width: values.dimensionWidth || 0,
          height: values.dimensionHeight || 0,
        };
      }

      if (isEdit) {
        await api.put(`/admin/products/${id}`, body);
        message.success("Товар обновлён");
      } else {
        const { data: resp } = await api.post("/admin/products", body);
        message.success("Товар создан");
        // Navigate to edit to allow image upload
        navigate(`/products/${resp.data.id}/edit`, { replace: true });
        return;
      }

      navigate("/products");
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

  const handleImageUpload = async (file: UploadFile) => {
    if (!id) return;
    const formData = new FormData();
    formData.append("file", file as unknown as Blob);
    setUploading(true);
    try {
      const { data: resp } = await api.post(
        `/admin/products/${id}/images`,
        formData,
        { headers: { "Content-Type": "multipart/form-data" } }
      );
      setImages((prev) => [...prev, resp.data]);
      message.success("Изображение загружено");
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message || "Ошибка загрузки"
      );
    } finally {
      setUploading(false);
    }
  };

  const handleSetMain = async (imageId: number) => {
    if (!id) return;
    try {
      await api.put(`/admin/products/${id}/images/${imageId}/main`);
      setImages((prev) =>
        prev.map((img) => ({ ...img, isMain: img.id === imageId }))
      );
      message.success("Главное фото обновлено");
    } catch {
      message.error("Ошибка при установке главного фото");
    }
  };

  const handleDeleteImage = async (imageId: number) => {
    try {
      await api.delete(`/admin/products/images/${imageId}`);
      setImages((prev) => prev.filter((img) => img.id !== imageId));
      message.success("Изображение удалено");
    } catch {
      message.error("Ошибка при удалении изображения");
    }
  };

  const flattenCategories = (
    cats: CategoryOption[],
    depth = 0
  ): { value: number; label: string }[] => {
    const result: { value: number; label: string }[] = [];
    for (const cat of cats) {
      const prefix = depth > 0 ? "\u2014 ".repeat(depth) : "";
      result.push({ value: cat.id, label: `${prefix}${cat.name}` });
      if (cat.children?.length) {
        result.push(...flattenCategories(cat.children, depth + 1));
      }
    }
    return result;
  };

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/products")}>
          Назад
        </Button>
        <Title level={4} style={{ margin: 0 }}>
          {isEdit ? "Редактировать товар" : "Новый товар"}
        </Title>
      </Space>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ isActive: true, isFeatured: false, stockQuantity: 0 }}
      >
        <Row gutter={24}>
          {/* Left column: main fields */}
          <Col xs={24} lg={16}>
            <Card title="Основная информация" style={{ marginBottom: 16 }}>
              <Form.Item
                name="name"
                label="Название"
                rules={[{ required: true, message: "Введите название" }]}
              >
                <Input placeholder="Фигурка дракона" />
              </Form.Item>

              <Form.Item
                name="slug"
                label="Slug (URL)"
                help="Оставьте пустым для автогенерации"
              >
                <Input placeholder="figurka-drakona" />
              </Form.Item>

              <Form.Item name="shortDescription" label="Краткое описание">
                <TextArea rows={2} placeholder="Короткое описание для каталога" />
              </Form.Item>

              <Form.Item name="description" label="Полное описание">
                <TextArea rows={6} placeholder="Подробное описание товара" />
              </Form.Item>
            </Card>

            <Card title="Характеристики" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col xs={12} md={8}>
                  <Form.Item name="material" label="Материал">
                    <Select
                      allowClear
                      placeholder="Выберите"
                      options={MATERIALS}
                    />
                  </Form.Item>
                </Col>
                <Col xs={12} md={8}>
                  <Form.Item name="weight" label="Вес (г)">
                    <InputNumber
                      min={0}
                      step={0.1}
                      style={{ width: "100%" }}
                      placeholder="350"
                    />
                  </Form.Item>
                </Col>
                <Col xs={12} md={8}>
                  <Form.Item name="printTime" label="Время печати (мин)">
                    <InputNumber
                      min={0}
                      style={{ width: "100%" }}
                      placeholder="240"
                    />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={16}>
                <Col xs={8}>
                  <Form.Item name="dimensionLength" label="Длина (см)">
                    <InputNumber
                      min={0}
                      step={0.1}
                      style={{ width: "100%" }}
                    />
                  </Form.Item>
                </Col>
                <Col xs={8}>
                  <Form.Item name="dimensionWidth" label="Ширина (см)">
                    <InputNumber
                      min={0}
                      step={0.1}
                      style={{ width: "100%" }}
                    />
                  </Form.Item>
                </Col>
                <Col xs={8}>
                  <Form.Item name="dimensionHeight" label="Высота (см)">
                    <InputNumber
                      min={0}
                      step={0.1}
                      style={{ width: "100%" }}
                    />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item name="sku" label="Артикул (SKU)">
                <Input placeholder="3DF-DRAGON-001" />
              </Form.Item>
            </Card>

            {/* Images — only in edit mode (need product ID) */}
            {isEdit && (
              <Card title="Изображения" style={{ marginBottom: 16 }}>
                <div style={{ display: "flex", flexWrap: "wrap", gap: 12 }}>
                  {images.map((img) => (
                    <div
                      key={img.id}
                      style={{
                        position: "relative",
                        width: 120,
                        height: 120,
                        border: img.isMain
                          ? "2px solid #1890ff"
                          : "1px solid #d9d9d9",
                        borderRadius: 8,
                        overflow: "hidden",
                      }}
                    >
                      <Image
                        src={img.urlThumbnail || img.url}
                        alt=""
                        width={120}
                        height={120}
                        style={{ objectFit: "cover" }}
                      />
                      <div
                        style={{
                          position: "absolute",
                          bottom: 0,
                          left: 0,
                          right: 0,
                          background: "rgba(0,0,0,0.5)",
                          display: "flex",
                          justifyContent: "center",
                          gap: 4,
                          padding: 4,
                        }}
                      >
                        <Button
                          type="text"
                          size="small"
                          icon={
                            img.isMain ? (
                              <StarFilled style={{ color: "#fadb14" }} />
                            ) : (
                              <StarOutlined style={{ color: "#fff" }} />
                            )
                          }
                          onClick={() => handleSetMain(img.id)}
                          title="Сделать главным"
                        />
                        <Button
                          type="text"
                          size="small"
                          danger
                          icon={<DeleteOutlined style={{ color: "#ff4d4f" }} />}
                          onClick={() => handleDeleteImage(img.id)}
                          title="Удалить"
                        />
                      </div>
                    </div>
                  ))}

                  <Upload
                    accept="image/jpeg,image/png,image/webp"
                    showUploadList={false}
                    beforeUpload={(file) => {
                      handleImageUpload(file as unknown as UploadFile);
                      return false;
                    }}
                    multiple
                  >
                    <div
                      style={{
                        width: 120,
                        height: 120,
                        border: "1px dashed #d9d9d9",
                        borderRadius: 8,
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        justifyContent: "center",
                        cursor: "pointer",
                        color: "#999",
                      }}
                    >
                      {uploading ? (
                        <Spin />
                      ) : (
                        <>
                          <PlusOutlined />
                          <div style={{ fontSize: 12, marginTop: 4 }}>
                            Загрузить
                          </div>
                        </>
                      )}
                    </div>
                  </Upload>
                </div>
                {!isEdit && (
                  <div style={{ color: "#999", marginTop: 8 }}>
                    Сохраните товар, чтобы загружать изображения
                  </div>
                )}
              </Card>
            )}
          </Col>

          {/* Right column: price, stock, settings */}
          <Col xs={24} lg={8}>
            <Card title="Цена и наличие" style={{ marginBottom: 16 }}>
              <Form.Item
                name="price"
                label="Цена (₽)"
                rules={[{ required: true, message: "Укажите цену" }]}
              >
                <InputNumber
                  min={0}
                  step={10}
                  style={{ width: "100%" }}
                  placeholder="990"
                />
              </Form.Item>

              <Form.Item name="oldPrice" label="Старая цена (₽)">
                <InputNumber
                  min={0}
                  step={10}
                  style={{ width: "100%" }}
                  placeholder="1490"
                />
              </Form.Item>

              <Form.Item name="stockQuantity" label="Остаток (шт)">
                <InputNumber min={0} style={{ width: "100%" }} />
              </Form.Item>
            </Card>

            <Card title="Настройки" style={{ marginBottom: 16 }}>
              <Form.Item name="categoryId" label="Категория">
                <Select
                  allowClear
                  placeholder="Выберите категорию"
                  options={flattenCategories(categories)}
                />
              </Form.Item>

              {isEdit && (
                <Form.Item
                  name="isActive"
                  label="Активен"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
              )}

              <Form.Item
                name="isFeatured"
                label="Показывать на главной"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Card>

            <Button
              type="primary"
              icon={<SaveOutlined />}
              size="large"
              block
              onClick={handleSubmit}
              loading={saving}
            >
              {isEdit ? "Сохранить" : "Создать товар"}
            </Button>
          </Col>
        </Row>
      </Form>
    </div>
  );
};

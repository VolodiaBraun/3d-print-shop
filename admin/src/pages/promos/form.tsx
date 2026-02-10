import { useState, useEffect } from "react";
import {
  Card,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  Button,
  DatePicker,
  Space,
  App,
  Spin,
} from "antd";
import { ArrowLeftOutlined } from "@ant-design/icons";
import { useParams, useNavigate } from "react-router-dom";
import dayjs from "dayjs";
import api from "../../lib/api";

interface PromoFormValues {
  code: string;
  description?: string;
  discountType: string;
  discountValue: number;
  minOrderAmount?: number;
  maxUses?: number;
  isActive: boolean;
  startsAt?: dayjs.Dayjs;
  expiresAt?: dayjs.Dayjs;
}

export const PromoForm = () => {
  const { message } = App.useApp();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm<PromoFormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const isEdit = !!id;

  useEffect(() => {
    if (!isEdit) return;
    setLoading(true);
    api
      .get(`/admin/promo-codes/${id}`)
      .then(({ data: resp }) => {
        const p = resp.data;
        form.setFieldsValue({
          code: p.code,
          description: p.description,
          discountType: p.discountType,
          discountValue: p.discountValue,
          minOrderAmount: p.minOrderAmount || undefined,
          maxUses: p.maxUses || undefined,
          isActive: p.isActive,
          startsAt: p.startsAt ? dayjs(p.startsAt) : undefined,
          expiresAt: p.expiresAt ? dayjs(p.expiresAt) : undefined,
        });
      })
      .catch(() => message.error("Не удалось загрузить промокод"))
      .finally(() => setLoading(false));
  }, [id, isEdit, form, message]);

  const handleSubmit = async (values: PromoFormValues) => {
    setSaving(true);
    try {
      const payload = {
        code: values.code.toUpperCase(),
        description: values.description || "",
        discountType: values.discountType,
        discountValue: values.discountValue,
        minOrderAmount: values.minOrderAmount || 0,
        maxUses: values.maxUses || 0,
        isActive: values.isActive ?? true,
        startsAt: values.startsAt
          ? values.startsAt.startOf("day").toISOString()
          : undefined,
        expiresAt: values.expiresAt
          ? values.expiresAt.endOf("day").toISOString()
          : undefined,
      };

      if (isEdit) {
        await api.put(`/admin/promo-codes/${id}`, payload);
        message.success("Промокод обновлён");
      } else {
        await api.post("/admin/promo-codes", payload);
        message.success("Промокод создан");
      }
      navigate("/promos");
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message || "Ошибка сохранения"
      );
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 64 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate("/promos")}
        style={{ marginBottom: 16 }}
      >
        Назад к промокодам
      </Button>

      <Card title={isEdit ? "Редактировать промокод" : "Создать промокод"}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ discountType: "percent", isActive: true }}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            name="code"
            label="Код промокода"
            rules={[
              { required: true, message: "Введите код" },
              { min: 3, message: "Минимум 3 символа" },
              { max: 50, message: "Максимум 50 символов" },
            ]}
          >
            <Input
              placeholder="SALE10"
              style={{ textTransform: "uppercase", fontFamily: "monospace" }}
            />
          </Form.Item>

          <Form.Item name="description" label="Описание">
            <Input.TextArea
              placeholder="Скидка 10% на первый заказ"
              rows={2}
            />
          </Form.Item>

          <Space size="middle" style={{ display: "flex" }}>
            <Form.Item
              name="discountType"
              label="Тип скидки"
              rules={[{ required: true }]}
            >
              <Select style={{ width: 180 }}>
                <Select.Option value="percent">Процент (%)</Select.Option>
                <Select.Option value="fixed">Фиксированная (₽)</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="discountValue"
              label="Значение скидки"
              rules={[
                { required: true, message: "Введите значение" },
                {
                  type: "number",
                  min: 0.01,
                  message: "Должно быть больше 0",
                },
              ]}
            >
              <InputNumber
                placeholder="10"
                style={{ width: 150 }}
                min={0.01}
              />
            </Form.Item>
          </Space>

          <Form.Item name="minOrderAmount" label="Минимальная сумма заказа (₽)">
            <InputNumber
              placeholder="0 — без ограничений"
              style={{ width: 250 }}
              min={0}
            />
          </Form.Item>

          <Form.Item name="maxUses" label="Лимит использований">
            <InputNumber
              placeholder="0 — без ограничений"
              style={{ width: 250 }}
              min={0}
            />
          </Form.Item>

          <Space size="middle" style={{ display: "flex" }}>
            <Form.Item name="startsAt" label="Начало действия">
              <DatePicker
                format="DD.MM.YYYY"
                placeholder="Без ограничений"
                style={{ width: 180 }}
              />
            </Form.Item>

            <Form.Item name="expiresAt" label="Окончание действия">
              <DatePicker
                format="DD.MM.YYYY"
                placeholder="Без ограничений"
                style={{ width: 180 }}
              />
            </Form.Item>
          </Space>

          <Form.Item
            name="isActive"
            label="Активен"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isEdit ? "Сохранить" : "Создать"}
              </Button>
              <Button onClick={() => navigate("/promos")}>Отмена</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

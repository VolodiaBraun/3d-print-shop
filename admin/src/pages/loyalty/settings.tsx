import { useState, useEffect } from "react";
import {
  Card,
  Form,
  InputNumber,
  Switch,
  Button,
  App,
  Spin,
  Statistic,
  Row,
  Col,
} from "antd";
import api from "../../lib/api";

interface LoyaltySettingsData {
  referrerBonusPercent: number;
  referralWelcomeBonus: number;
  isActive: boolean;
}

export const LoyaltySettings = () => {
  const { message } = App.useApp();
  const [form] = Form.useForm<LoyaltySettingsData>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<LoyaltySettingsData | null>(null);

  useEffect(() => {
    api
      .get("/admin/loyalty/settings")
      .then(({ data: resp }) => {
        const s = resp.data;
        setSettings(s);
        form.setFieldsValue({
          referrerBonusPercent: s.referrerBonusPercent,
          referralWelcomeBonus: s.referralWelcomeBonus,
          isActive: s.isActive,
        });
      })
      .catch(() => message.error("Не удалось загрузить настройки"))
      .finally(() => setLoading(false));
  }, [form, message]);

  const handleSubmit = async (values: LoyaltySettingsData) => {
    setSaving(true);
    try {
      const { data: resp } = await api.put("/admin/loyalty/settings", values);
      setSettings(resp.data);
      message.success("Настройки сохранены");
    } catch {
      message.error("Ошибка сохранения");
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
      <h2 style={{ marginBottom: 24 }}>Реферальная программа</h2>

      {settings && (
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={8}>
            <Card>
              <Statistic
                title="Бонус рефереру"
                value={settings.referrerBonusPercent}
                suffix="%"
                precision={2}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="Приветственный бонус"
                value={settings.referralWelcomeBonus}
                suffix="₽"
                precision={0}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="Статус"
                value={settings.isActive ? "Активна" : "Неактивна"}
                valueStyle={{
                  color: settings.isActive ? "#52c41a" : "#ff4d4f",
                }}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Card title="Настройки">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          style={{ maxWidth: 500 }}
        >
          <Form.Item
            name="referrerBonusPercent"
            label="Процент бонуса для реферера (%)"
            tooltip="Какой процент от суммы заказа реферала получает реферер при доставке"
            rules={[
              { required: true, message: "Введите значение" },
              {
                type: "number",
                min: 0,
                max: 100,
                message: "От 0 до 100%",
              },
            ]}
          >
            <InputNumber
              style={{ width: 200 }}
              min={0}
              max={100}
              step={0.5}
              addonAfter="%"
            />
          </Form.Item>

          <Form.Item
            name="referralWelcomeBonus"
            label="Приветственный бонус для реферала (₽)"
            tooltip="Сумма бонусов, которую получает новый пользователь при вводе реферального кода"
            rules={[
              { required: true, message: "Введите значение" },
              {
                type: "number",
                min: 0,
                message: "Не может быть отрицательным",
              },
            ]}
          >
            <InputNumber
              style={{ width: 200 }}
              min={0}
              step={10}
              addonAfter="₽"
            />
          </Form.Item>

          <Form.Item
            name="isActive"
            label="Программа активна"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={saving}>
              Сохранить
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

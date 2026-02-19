import { useState, useEffect } from "react";
import { Card, Form, Input, Button, App, Spin, Tabs } from "antd";
import api from "../../lib/api";

interface HeroData {
  title: string;
  titleAccent: string;
  subtitle: string;
  ctaPrimary: string;
  ctaPrimaryLink: string;
  ctaSecondary: string;
  ctaSecondaryLink: string;
}

interface FooterData {
  description: string;
  telegram: string;
  email: string;
  phone: string;
  address: string;
}

export const ContentEditor = () => {
  const { message } = App.useApp();
  const [heroForm] = Form.useForm<HeroData>();
  const [footerForm] = Form.useForm<FooterData>();
  const [loading, setLoading] = useState(true);
  const [savingHero, setSavingHero] = useState(false);
  const [savingFooter, setSavingFooter] = useState(false);

  useEffect(() => {
    Promise.all([
      api.get("/content/hero").catch(() => null),
      api.get("/content/footer").catch(() => null),
    ])
      .then(([heroResp, footerResp]) => {
        if (heroResp?.data) {
          heroForm.setFieldsValue(heroResp.data);
        }
        if (footerResp?.data) {
          footerForm.setFieldsValue(footerResp.data);
        }
      })
      .finally(() => setLoading(false));
  }, [heroForm, footerForm]);

  const saveHero = async (values: HeroData) => {
    setSavingHero(true);
    try {
      await api.put("/admin/content/hero", { data: values });
      message.success("Главный экран обновлён");
    } catch {
      message.error("Ошибка сохранения");
    } finally {
      setSavingHero(false);
    }
  };

  const saveFooter = async (values: FooterData) => {
    setSavingFooter(true);
    try {
      await api.put("/admin/content/footer", { data: values });
      message.success("Подвал обновлён");
    } catch {
      message.error("Ошибка сохранения");
    } finally {
      setSavingFooter(false);
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
      <h2 style={{ marginBottom: 24 }}>Управление контентом</h2>

      <Tabs
        items={[
          {
            key: "hero",
            label: "Главный экран",
            children: (
              <Card>
                <Form
                  form={heroForm}
                  layout="vertical"
                  onFinish={saveHero}
                  style={{ maxWidth: 600 }}
                >
                  <Form.Item
                    name="title"
                    label="Заголовок"
                    rules={[{ required: true, message: "Введите заголовок" }]}
                  >
                    <Input placeholder="3D-печатные изделия" />
                  </Form.Item>

                  <Form.Item
                    name="titleAccent"
                    label="Акцентная часть заголовка"
                    rules={[{ required: true, message: "Введите акцент" }]}
                  >
                    <Input placeholder="премиум качества" />
                  </Form.Item>

                  <Form.Item
                    name="subtitle"
                    label="Подзаголовок"
                    rules={[
                      { required: true, message: "Введите подзаголовок" },
                    ]}
                  >
                    <Input.TextArea
                      rows={3}
                      placeholder="Фигурки, модели и аксессуары..."
                    />
                  </Form.Item>

                  <Form.Item name="ctaPrimary" label="Основная кнопка — текст">
                    <Input placeholder="Смотреть каталог" />
                  </Form.Item>

                  <Form.Item
                    name="ctaPrimaryLink"
                    label="Основная кнопка — ссылка"
                  >
                    <Input placeholder="/catalog" />
                  </Form.Item>

                  <Form.Item
                    name="ctaSecondary"
                    label="Вторая кнопка — текст"
                  >
                    <Input placeholder="Новинки" />
                  </Form.Item>

                  <Form.Item
                    name="ctaSecondaryLink"
                    label="Вторая кнопка — ссылка"
                  >
                    <Input placeholder="/catalog?sort=newest" />
                  </Form.Item>

                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={savingHero}
                    >
                      Сохранить
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            ),
          },
          {
            key: "footer",
            label: "Подвал",
            children: (
              <Card>
                <Form
                  form={footerForm}
                  layout="vertical"
                  onFinish={saveFooter}
                  style={{ maxWidth: 600 }}
                >
                  <Form.Item
                    name="description"
                    label="Описание компании"
                    rules={[{ required: true, message: "Введите описание" }]}
                  >
                    <Input.TextArea
                      rows={3}
                      placeholder="3D-печатные изделия премиального качества..."
                    />
                  </Form.Item>

                  <Form.Item name="telegram" label="Telegram">
                    <Input placeholder="@avangard3d" />
                  </Form.Item>

                  <Form.Item name="email" label="Email">
                    <Input placeholder="info@avangard-print.ru" />
                  </Form.Item>

                  <Form.Item name="phone" label="Телефон">
                    <Input placeholder="+7 (999) 123-45-67" />
                  </Form.Item>

                  <Form.Item name="address" label="Адрес">
                    <Input placeholder="г. Москва, ул. Примерная, д. 1" />
                  </Form.Item>

                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={savingFooter}
                    >
                      Сохранить
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            ),
          },
        ]}
      />
    </div>
  );
};

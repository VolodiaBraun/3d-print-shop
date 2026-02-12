import { useState, useEffect, useCallback } from "react";
import {
  Typography,
  Card,
  Row,
  Col,
  Statistic,
  Select,
  Table,
  Tag,
  Spin,
  Alert,
  Space,
} from "antd";
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  ShoppingCartOutlined,
  DollarOutlined,
  UserAddOutlined,
  CreditCardOutlined,
  WarningOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import api from "../lib/api";

const { Title, Text } = Typography;

interface PeriodMetric {
  today: number;
  week: number;
  month: number;
  prevWeek: number;
  prevMonth: number;
  weekChange: number;
  monthChange: number;
}

interface TopProduct {
  productId: number;
  name: string;
  slug: string;
  totalSold: number;
  revenue: number;
  imageUrl?: string;
}

interface LowStockProduct {
  id: number;
  name: string;
  slug: string;
  stockQuantity: number;
}

interface PendingOrder {
  id: number;
  orderNumber: string;
  totalPrice: number;
  createdAt: string;
  hoursPending: number;
}

interface DashboardMetrics {
  revenue: PeriodMetric;
  ordersCount: PeriodMetric;
  avgCheck: PeriodMetric;
  newCustomers: PeriodMetric;
  topProducts: TopProduct[];
  lowStock: LowStockProduct[];
  pendingOrders: PendingOrder[];
}

interface ChartDataPoint {
  date: string;
  revenue: number;
  ordersCount: number;
}

const formatPrice = (value: number) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);

const formatDate = (dateStr: string) => {
  const d = new Date(dateStr);
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
};

const MetricCard = ({
  title,
  metric,
  icon,
  formatter = (v: number) => String(Math.round(v)),
  color,
}: {
  title: string;
  metric: PeriodMetric;
  icon: React.ReactNode;
  formatter?: (v: number) => string;
  color: string;
}) => {
  const weekUp = metric.weekChange >= 0;
  const monthUp = metric.monthChange >= 0;

  return (
    <Card>
      <Statistic
        title={title}
        value={metric.today}
        formatter={(val) => formatter(val as number)}
        prefix={icon}
        valueStyle={{ color }}
      />
      <div style={{ marginTop: 12, fontSize: 13 }}>
        <Space direction="vertical" size={2} style={{ width: "100%" }}>
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <Text type="secondary">Неделя: {formatter(metric.week)}</Text>
            <Text style={{ color: weekUp ? "#3f8600" : "#cf1322" }}>
              {weekUp ? <ArrowUpOutlined /> : <ArrowDownOutlined />}{" "}
              {Math.abs(Math.round(metric.weekChange))}%
            </Text>
          </div>
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <Text type="secondary">Месяц: {formatter(metric.month)}</Text>
            <Text style={{ color: monthUp ? "#3f8600" : "#cf1322" }}>
              {monthUp ? <ArrowUpOutlined /> : <ArrowDownOutlined />}{" "}
              {Math.abs(Math.round(metric.monthChange))}%
            </Text>
          </div>
        </Space>
      </div>
    </Card>
  );
};

export const DashboardPage = () => {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [chartPeriod, setChartPeriod] = useState("month");
  const [loading, setLoading] = useState(true);
  const [chartLoading, setChartLoading] = useState(false);

  useEffect(() => {
    api
      .get("/admin/analytics/dashboard")
      .then((res) => {
        const m = res.data.data;
        if (m) {
          m.topProducts = m.topProducts || [];
          m.lowStock = m.lowStock || [];
          m.pendingOrders = m.pendingOrders || [];
        }
        setMetrics(m);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const loadChart = useCallback((period: string) => {
    setChartLoading(true);
    api
      .get("/admin/analytics/chart", { params: { period } })
      .then((res) => setChartData(res.data.data || []))
      .catch(() => {})
      .finally(() => setChartLoading(false));
  }, []);

  useEffect(() => {
    loadChart(chartPeriod);
  }, [chartPeriod, loadChart]);

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!metrics) {
    return <Alert type="error" message="Не удалось загрузить данные" />;
  }

  const topProductColumns = [
    {
      title: "Товар",
      dataIndex: "name",
      key: "name",
      ellipsis: true,
    },
    {
      title: "Продано",
      dataIndex: "totalSold",
      key: "totalSold",
      width: 90,
      align: "right" as const,
    },
    {
      title: "Выручка",
      dataIndex: "revenue",
      key: "revenue",
      width: 120,
      align: "right" as const,
      render: (v: number) => formatPrice(v),
    },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        Аналитика
      </Title>

      {/* Metric cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <MetricCard
            title="Выручка (сегодня)"
            metric={metrics.revenue}
            icon={<DollarOutlined />}
            formatter={formatPrice}
            color="#1677ff"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <MetricCard
            title="Заказы (сегодня)"
            metric={metrics.ordersCount}
            icon={<ShoppingCartOutlined />}
            color="#722ed1"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <MetricCard
            title="Средний чек (сегодня)"
            metric={metrics.avgCheck}
            icon={<CreditCardOutlined />}
            formatter={formatPrice}
            color="#13c2c2"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <MetricCard
            title="Новые клиенты (сегодня)"
            metric={metrics.newCustomers}
            icon={<UserAddOutlined />}
            color="#52c41a"
          />
        </Col>
      </Row>

      {/* Chart */}
      <Card
        title="График продаж"
        extra={
          <Select
            value={chartPeriod}
            onChange={setChartPeriod}
            style={{ width: 140 }}
            options={[
              { value: "week", label: "Неделя" },
              { value: "month", label: "Месяц" },
              { value: "quarter", label: "Квартал" },
              { value: "year", label: "Год" },
            ]}
          />
        }
        style={{ marginBottom: 24 }}
      >
        {chartLoading ? (
          <div style={{ textAlign: "center", padding: 40 }}>
            <Spin />
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={320}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" tickFormatter={formatDate} fontSize={12} />
              <YAxis
                yAxisId="revenue"
                tickFormatter={(v) => `${Math.round(v / 1000)}k`}
                fontSize={12}
              />
              <YAxis
                yAxisId="orders"
                orientation="right"
                fontSize={12}
              />
              <Tooltip
                labelFormatter={(label) => {
                  const d = new Date(label);
                  return d.toLocaleDateString("ru-RU", {
                    day: "numeric",
                    month: "long",
                    year: "numeric",
                  });
                }}
                formatter={(value: number | undefined, name: string) => {
                  if (value == null) return [0, name];
                  if (name === "Выручка") return [formatPrice(value), name];
                  return [value, name];
                }}
              />
              <Legend />
              <Line
                yAxisId="revenue"
                type="monotone"
                dataKey="revenue"
                stroke="#1677ff"
                strokeWidth={2}
                dot={false}
                name="Выручка"
              />
              <Line
                yAxisId="orders"
                type="monotone"
                dataKey="ordersCount"
                stroke="#8c8c8c"
                strokeWidth={2}
                dot={false}
                name="Заказы"
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </Card>

      {/* Bottom section: top products + alerts */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="Топ-5 товаров">
            <Table
              dataSource={metrics.topProducts}
              columns={topProductColumns}
              rowKey="productId"
              pagination={false}
              size="small"
              locale={{ emptyText: "Нет данных" }}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Требует внимания">
            {metrics.lowStock.length === 0 &&
            metrics.pendingOrders.length === 0 ? (
              <Text type="secondary" style={{ display: "block", textAlign: "center", padding: 20 }}>
                Нет проблем
              </Text>
            ) : (
              <Space direction="vertical" size={8} style={{ width: "100%" }}>
                {metrics.lowStock.map((p) => (
                  <Alert
                    key={`stock-${p.id}`}
                    type="warning"
                    showIcon
                    icon={<WarningOutlined />}
                    message={
                      <span>
                        <b>{p.name}</b> — осталось{" "}
                        <Tag color="orange">{p.stockQuantity} шт.</Tag>
                      </span>
                    }
                  />
                ))}
                {metrics.pendingOrders.map((o) => (
                  <Alert
                    key={`order-${o.id}`}
                    type="error"
                    showIcon
                    icon={<ClockCircleOutlined />}
                    message={
                      <span>
                        Заказ <b>{o.orderNumber}</b> — ждёт{" "}
                        <Tag color="red">{Math.round(o.hoursPending)}ч</Tag>{" "}
                        ({formatPrice(o.totalPrice)})
                      </span>
                    }
                  />
                ))}
              </Space>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

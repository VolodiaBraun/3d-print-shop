import { Typography, Card } from "antd";

const { Title, Paragraph } = Typography;

export const DashboardPage = () => {
  return (
    <div>
      <Title level={2}>Интернет-магазин 3D-печати</Title>
      <Card>
        <Paragraph style={{ fontSize: 16, textAlign: "center", padding: "40px 0" }}>
          Dashboard coming soon
        </Paragraph>
      </Card>
    </div>
  );
};

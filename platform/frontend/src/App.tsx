import { Layout, Menu, Spin, Typography, Button, Avatar, Space } from "antd";
import { CloudServerOutlined, LogoutOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "./context/AuthContext";
import LoginPage from "./pages/LoginPage";
import ProjectsPage from "./pages/ProjectsPage";
import ProjectPage from "./pages/ProjectPage";
import ScanPage from "./pages/ScanPage";

const { Header, Content } = Layout;

function Protected({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  const location = useLocation();
  if (loading) {
    return <div className="center-screen"><Spin size="large" /></div>;
  }
  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}

function Shell({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  return (
    <Layout className="app-shell">
      <Header className="app-header">
        <button className="brand" onClick={() => navigate("/")} aria-label="返回项目首页">
          <span className="brand-mark"><SafetyCertificateOutlined /></span>
          <span>DevSecurity</span>
          <small>SECURE SDLC</small>
        </button>
        <Menu
          mode="horizontal"
          selectedKeys={["projects"]}
          items={[{ key: "projects", icon: <CloudServerOutlined />, label: "项目" }]}
          onClick={() => navigate("/")}
          className="main-menu"
        />
        <Space>
          <Avatar>{user?.displayName?.slice(0, 1)}</Avatar>
          <Typography.Text className="header-user">{user?.displayName}</Typography.Text>
          <Button type="text" icon={<LogoutOutlined />} onClick={logout} aria-label="退出登录" />
        </Space>
      </Header>
      <Content className="app-content">{children}</Content>
    </Layout>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<Protected><Shell><ProjectsPage /></Shell></Protected>} />
      <Route path="/projects/:projectId" element={<Protected><Shell><ProjectPage /></Shell></Protected>} />
      <Route path="/scans/:scanId" element={<Protected><Shell><ScanPage /></Shell></Protected>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

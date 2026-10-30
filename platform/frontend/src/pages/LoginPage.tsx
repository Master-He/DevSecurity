import { useState } from "react";
import { Alert, Button, Card, Form, Input, Typography } from "antd";
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from "@ant-design/icons";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export default function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  if (user) {
    return <Navigate to="/" replace />;
  }

  async function submit(values: { username: string; password: string }) {
    setSubmitting(true);
    setError("");
    try {
      await login(values.username, values.password);
      const target = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ?? "/";
      navigate(target, { replace: true });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "登录失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-story">
        <div className="story-grid" />
        <div className="story-content">
          <div className="story-kicker">DEVSECURITY / CONTROL PLANE</div>
          <h1>把代码风险<br />变成可执行结论</h1>
          <p>统一编排 SCA、SAST、DAST 与 IAST，让安全结果进入每一位研发的工作流。</p>
          <div className="story-metrics">
            <div><strong>SCA</strong><span>组件与依赖风险</span></div>
            <div><strong>SAST</strong><span>源码缺陷定位</span></div>
            <div><strong>API</strong><span>DevOps 原生集成</span></div>
          </div>
        </div>
      </section>
      <section className="login-panel">
        <Card className="login-card" bordered={false}>
          <div className="login-icon"><SafetyCertificateOutlined /></div>
          <Typography.Title level={2}>登录安全平台</Typography.Title>
          <Typography.Paragraph type="secondary">
            使用平台账号进入项目与扫描控制台。
          </Typography.Paragraph>
          {error && <Alert type="error" message={error} showIcon className="form-alert" />}
          <Form layout="vertical" onFinish={submit} requiredMark={false}>
            <Form.Item name="username" label="用户名" rules={[{ required: true, message: "请输入用户名" }]}>
              <Input size="large" prefix={<UserOutlined />} autoComplete="username" />
            </Form.Item>
            <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
              <Input.Password size="large" prefix={<LockOutlined />} autoComplete="current-password" />
            </Form.Item>
            <Button block size="large" type="primary" htmlType="submit" loading={submitting}>
              进入平台
            </Button>
          </Form>
          <p className="login-hint">初始管理员密码由部署脚本生成，不在代码中保存固定密码。</p>
        </Card>
      </section>
    </main>
  );
}

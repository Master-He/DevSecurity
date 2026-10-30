import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Card, Empty, Form, Input, Modal, Skeleton, Typography, message } from "antd";
import { ArrowRightOutlined, FolderOpenOutlined, PlusOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";

export default function ProjectsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const projects = useQuery({ queryKey: ["projects"], queryFn: api.projects });
  const create = useMutation({
    mutationFn: ({ name, description }: { name: string; description: string }) =>
      api.createProject(name, description),
    onSuccess: (project) => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      setOpen(false);
      form.resetFields();
      navigate(`/projects/${project.id}`);
    },
    onError: (error) => message.error(error.message)
  });

  return (
    <div className="page">
      <header className="page-heading">
        <div>
          <span className="eyebrow">PROJECTS / OVERVIEW</span>
          <Typography.Title>安全项目</Typography.Title>
          <Typography.Paragraph type="secondary">
            每个项目聚合源码扫描、组件风险与版本修复结论。
          </Typography.Paragraph>
        </div>
        <Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          创建项目
        </Button>
      </header>

      {projects.isError && <Alert type="error" showIcon message={projects.error.message} />}
      {projects.isLoading ? (
        <div className="project-grid"><Skeleton active /><Skeleton active /><Skeleton active /></div>
      ) : projects.data?.length ? (
        <div className="project-grid">
          {projects.data.map((project, index) => (
            <Card
              key={project.id}
              className="project-card"
              hoverable
              onClick={() => navigate(`/projects/${project.id}`)}
            >
              <div className="project-index">{String(index + 1).padStart(2, "0")}</div>
              <div className="project-icon"><FolderOpenOutlined /></div>
              <Typography.Title level={3}>{project.name}</Typography.Title>
              <Typography.Paragraph ellipsis={{ rows: 2 }}>
                {project.description || "暂无项目描述"}
              </Typography.Paragraph>
              <div className="project-footer">
                <span>{project.ownerUsername}</span>
                <ArrowRightOutlined />
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="empty-card">
          <Empty description="还没有安全项目">
            <Button type="primary" onClick={() => setOpen(true)}>创建第一个项目</Button>
          </Empty>
        </Card>
      )}

      <Modal
        title="创建安全项目"
        open={open}
        okText="创建"
        cancelText="取消"
        confirmLoading={create.isPending}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
      >
        <Form layout="vertical" form={form} onFinish={(values) => create.mutate(values)}>
          <Form.Item name="name" label="项目名称" rules={[{ required: true, message: "请输入项目名称" }]}>
            <Input placeholder="例如：订单中心" maxLength={120} />
          </Form.Item>
          <Form.Item name="description" label="项目说明">
            <Input.TextArea placeholder="业务、负责人或扫描范围" rows={3} maxLength={1000} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

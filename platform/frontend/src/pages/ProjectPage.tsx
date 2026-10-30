import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert, Button, Card, Empty, Form, Input, Modal, Space, Table, Tabs, Tag, Typography, Upload, message
} from "antd";
import type { UploadFile } from "antd";
import { ArrowLeftOutlined, CloudUploadOutlined, CodeOutlined, PlusOutlined } from "@ant-design/icons";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import type { Scan } from "../types";
import ScanStatusTag from "../components/ScanStatusTag";
import { formatTime } from "../utils/format";

export default function ProjectPage() {
  const { projectId = "" } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState("upload");
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [form] = Form.useForm();
  const project = useQuery({ queryKey: ["project", projectId], queryFn: () => api.project(projectId) });
  const scans = useQuery({
    queryKey: ["scans", projectId],
    queryFn: () => api.scans(projectId),
    refetchInterval: 4000
  });
  const createGit = useMutation({
    mutationFn: (values: { gitUrl: string; revision?: string }) =>
      api.createGitScan(projectId, values.gitUrl, values.revision ?? ""),
    onSuccess: finishCreation,
    onError: (error) => message.error(error.message)
  });
  const createUpload = useMutation({
    mutationFn: (file: File) => api.createUploadScan(projectId, file),
    onSuccess: finishCreation,
    onError: (error) => message.error(error.message)
  });

  function finishCreation(scan: Scan) {
    void queryClient.invalidateQueries({ queryKey: ["scans", projectId] });
    setOpen(false);
    setFileList([]);
    form.resetFields();
    navigate(`/scans/${scan.id}`);
  }

  function submitUpload() {
    const file = fileList[0]?.originFileObj;
    if (!file) {
      message.error("请选择 Java 源码 ZIP");
      return;
    }
    createUpload.mutate(file);
  }

  const columns = [
    {
      title: "任务",
      dataIndex: "id",
      render: (value: string) => <Typography.Text code>{value.slice(0, 8)}</Typography.Text>
    },
    { title: "来源", dataIndex: "sourceType", render: (value: string) => value === "GIT" ? "Git 链接" : "ZIP 上传" },
    { title: "状态", dataIndex: "status", render: (value: Scan["status"]) => <ScanStatusTag status={value} /> },
    { title: "组件", dataIndex: "componentCount" },
    {
      title: "漏洞",
      dataIndex: "findingCount",
      render: (value: number, scan: Scan) => scan.status === "SUCCEEDED"
        ? <Typography.Text strong={value > 0} type={value > 0 ? "danger" : undefined}>{value}</Typography.Text>
        : "—"
    },
    { title: "创建时间", dataIndex: "createdAt", render: formatTime },
    {
      title: "",
      render: (_: unknown, scan: Scan) => <Button type="link" onClick={() => navigate(`/scans/${scan.id}`)}>查看</Button>
    }
  ];

  return (
    <div className="page">
      <Button className="back-link" type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate("/")}>
        返回项目
      </Button>
      <header className="page-heading project-heading">
        <div>
          <span className="eyebrow">PROJECT / SECURITY WORKSPACE</span>
          <Typography.Title>{project.data?.name ?? "项目"}</Typography.Title>
          <Typography.Paragraph type="secondary">
            {project.data?.description || "暂无项目描述"}
          </Typography.Paragraph>
        </div>
        <Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          新建扫描
        </Button>
      </header>

      {project.isError && <Alert type="error" showIcon message={project.error.message} />}
      <div className="stat-strip">
        <div><span>扫描任务</span><strong>{scans.data?.length ?? 0}</strong></div>
        <div><span>最近组件</span><strong>{scans.data?.[0]?.componentCount ?? 0}</strong></div>
        <div><span>最近漏洞</span><strong className={scans.data?.[0]?.findingCount ? "danger-text" : ""}>{scans.data?.[0]?.findingCount ?? 0}</strong></div>
        <div><span>引擎</span><strong className="stat-small">SCA / Maven</strong></div>
      </div>
      <Card className="table-card" title="扫描记录" extra={<span className="muted">源码保留 3 天 · 结论保留 180 天</span>}>
        {scans.isError ? (
          <Alert type="error" showIcon message={scans.error.message} />
        ) : scans.data?.length ? (
          <Table
            rowKey="id"
            columns={columns}
            dataSource={scans.data}
            pagination={false}
            onRow={(scan) => ({ onClick: () => navigate(`/scans/${scan.id}`), className: "clickable-row" })}
          />
        ) : <Empty description="尚未执行扫描" />}
      </Card>

      <Modal
        title="新建 SCA 扫描"
        open={open}
        width={680}
        footer={null}
        onCancel={() => setOpen(false)}
      >
        <Tabs
          activeKey={mode}
          onChange={setMode}
          items={[
            {
              key: "upload",
              label: <span><CloudUploadOutlined /> 上传 ZIP</span>,
              children: (
                <Space direction="vertical" size="large" style={{ width: "100%" }}>
                  <Upload.Dragger
                    accept=".zip"
                    maxCount={1}
                    beforeUpload={() => false}
                    fileList={fileList}
                    onChange={({ fileList: next }) => setFileList(next)}
                  >
                    <p className="ant-upload-drag-icon"><CloudUploadOutlined /></p>
                    <p className="ant-upload-text">点击或拖入 Java 源码 ZIP</p>
                    <p className="ant-upload-hint">必须包含 pom.xml，最大 100 MB</p>
                  </Upload.Dragger>
                  <Button type="primary" block loading={createUpload.isPending} onClick={submitUpload}>
                    创建异步扫描
                  </Button>
                </Space>
              )
            },
            {
              key: "git",
              label: <span><CodeOutlined /> Git 链接</span>,
              children: (
                <Form layout="vertical" form={form} onFinish={(values) => createGit.mutate(values)}>
                  <Alert
                    type="info"
                    showIcon
                    message="仅允许管理员配置的 Git 主机，源码会在异步任务中拉取。"
                    className="form-alert"
                  />
                  <Form.Item name="gitUrl" label="Git HTTP(S) 地址" rules={[{ required: true, message: "请输入 Git 地址" }]}>
                    <Input placeholder="https://git.example.com/team/repo.git" />
                  </Form.Item>
                  <Form.Item name="revision" label="分支或 Tag（可选）">
                    <Input placeholder="main" />
                  </Form.Item>
                  <Button type="primary" htmlType="submit" block loading={createGit.isPending}>
                    拉取并创建扫描
                  </Button>
                </Form>
              )
            }
          ]}
        />
      </Modal>
    </div>
  );
}

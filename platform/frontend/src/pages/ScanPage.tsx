import { useQuery } from "@tanstack/react-query";
import { Alert, Button, Card, Descriptions, Empty, Space, Table, Tabs, Tag, Typography } from "antd";
import { ArrowLeftOutlined, ReloadOutlined } from "@ant-design/icons";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import type { Component, Finding } from "../types";
import ScanStatusTag from "../components/ScanStatusTag";
import { formatTime } from "../utils/format";

const severityColor: Record<string, string> = {
  CRITICAL: "magenta",
  HIGH: "red",
  MEDIUM: "orange",
  LOW: "blue",
  INFO: "default"
};

export default function ScanPage() {
  const { scanId = "" } = useParams();
  const navigate = useNavigate();
  const scan = useQuery({
    queryKey: ["scan", scanId],
    queryFn: () => api.scan(scanId),
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === "QUEUED" || status === "RUNNING" ? 2500 : false;
    }
  });
  const completed = scan.data?.status === "SUCCEEDED";
  const isSast = scan.data?.scanType === "SAST";
  const findings = useQuery({
    queryKey: ["findings", scanId],
    queryFn: () => api.findings(scanId),
    enabled: completed
  });
  const components = useQuery({
    queryKey: ["components", scanId],
    queryFn: () => api.components(scanId),
    enabled: completed && !isSast
  });

  const scaFindingColumns = [
    {
      title: "风险",
      dataIndex: "severity",
      width: 110,
      render: (value: string) => <Tag color={severityColor[value] ?? "default"}>{value}</Tag>
    },
    {
      title: "漏洞",
      dataIndex: "vulnerabilityId",
      width: 170,
      render: (value: string) => <Typography.Text strong>{value}</Typography.Text>
    },
    { title: "标题", dataIndex: "title" },
    { title: "组件", dataIndex: "packageName", width: 240 },
    { title: "当前版本", dataIndex: "affectedVersion", width: 130 },
    { title: "修复版本", dataIndex: "fixedVersions", width: 140, render: (value: string) => value || "—" },
    {
      title: "依赖路径",
      dataIndex: "dependencyPath",
      width: 360,
      render: (value: string) => <Typography.Text className="path-cell" ellipsis={{ tooltip: value }}>{value}</Typography.Text>
    }
  ];
  const sastFindingColumns = [
    {
      title: "风险",
      dataIndex: "severity",
      width: 100,
      render: (value: string) => <Tag color={severityColor[value] ?? "default"}>{value}</Tag>
    },
    {
      title: "规则",
      dataIndex: "ruleId",
      width: 150,
      render: (value: string) => <Typography.Text code>{value}</Typography.Text>
    },
    { title: "标题", dataIndex: "title", width: 220 },
    {
      title: "置信度",
      dataIndex: "confidence",
      width: 100,
      render: (value: string) => <Tag color={value === "HIGH" ? "blue" : "default"}>{value || "—"}</Tag>
    },
    { title: "CWE", dataIndex: "cwe", width: 110 },
    {
      title: "位置",
      width: 260,
      render: (_: unknown, finding: Finding) => (
        <Typography.Text className="path-cell" ellipsis={{ tooltip: `${finding.filePath}:${finding.startLine}` }}>
          {finding.filePath}:{finding.startLine}
        </Typography.Text>
      )
    },
    {
      title: "代码",
      dataIndex: "codeSnippet",
      width: 360,
      render: (value: string) => (
        <Typography.Text code className="code-snippet" ellipsis={{ tooltip: value }}>{value}</Typography.Text>
      )
    },
    {
      title: "修复建议",
      dataIndex: "recommendation",
      width: 360,
      render: (value: string) => <Typography.Text className="path-cell">{value}</Typography.Text>
    }
  ];
  const componentColumns = [
    { title: "Group", dataIndex: "groupId" },
    { title: "Artifact", dataIndex: "artifactId" },
    { title: "版本", dataIndex: "version", width: 150 },
    {
      title: "关系",
      dataIndex: "direct",
      width: 90,
      render: (value: boolean) => <Tag color={value ? "blue" : "default"}>{value ? "直接" : "传递"}</Tag>
    },
    { title: "Scope", dataIndex: "scope", width: 100 },
    {
      title: "依赖路径",
      dataIndex: "dependencyPath",
      render: (value: string) => <Typography.Text className="path-cell" ellipsis={{ tooltip: value }}>{value}</Typography.Text>
    },
    { title: "PURL", dataIndex: "purl", render: (value: string) => <Typography.Text code>{value}</Typography.Text> }
  ];

  return (
    <div className="page">
      <Button
        className="back-link"
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => scan.data && navigate(`/projects/${scan.data.projectId}`)}
      >
        返回项目
      </Button>
      {scan.isError && <Alert type="error" showIcon message={scan.error.message} />}
      <header className="page-heading scan-heading">
        <div>
          <span className="eyebrow">SCAN / {isSast ? "SAST JOERN" : "SCA MAVEN"}</span>
          <Space align="center">
            <Typography.Title>扫描 {scanId.slice(0, 8)}</Typography.Title>
            {scan.data && <ScanStatusTag status={scan.data.status} />}
          </Space>
          <Typography.Paragraph type="secondary">
            {scan.data?.sourceType === "GIT" ? "Git 仓库" : "ZIP 源码包"} · {scan.data?.sourceRef || "本地上传"}
          </Typography.Paragraph>
        </div>
        <Button icon={<ReloadOutlined />} onClick={() => void scan.refetch()} loading={scan.isFetching}>
          刷新
        </Button>
      </header>

      {scan.data?.status === "FAILED" && (
        <Alert
          type="error"
          showIcon
          message={`扫描失败：${scan.data.failureCode}`}
          description={scan.data.failureMessage}
          className="scan-alert"
        />
      )}
      {(scan.data?.status === "QUEUED" || scan.data?.status === "RUNNING") && (
        <Alert type="info" showIcon message="扫描正在异步执行，页面会自动刷新。" className="scan-alert" />
      )}

      <div className="stat-strip">
        {!isSast && <div><span>组件</span><strong>{scan.data?.componentCount ?? 0}</strong></div>}
        <div><span>问题</span><strong className={scan.data?.findingCount ? "danger-text" : ""}>{scan.data?.findingCount ?? 0}</strong></div>
        <div><span>{isSast ? "规则引擎" : "解析方式"}</span><strong className="stat-small">{scan.data?.sourceMode || "等待执行"}</strong></div>
        <div><span>源码删除</span><strong className="stat-small">{scan.data?.sourceDeleted ? "已清理" : formatTime(scan.data?.sourceExpiresAt)}</strong></div>
        {isSast && <div><span>结论到期</span><strong className="stat-small">{formatTime(scan.data?.resultExpiresAt)}</strong></div>}
      </div>

      <Card className="table-card">
        <Tabs
          items={[
            {
              key: "findings",
              label: `${isSast ? "源码风险" : "漏洞风险"} ${scan.data?.findingCount ?? 0}`,
              children: findings.data?.length ? (
                <Table
                  rowKey="id"
                  columns={isSast ? sastFindingColumns : scaFindingColumns}
                  dataSource={findings.data}
                  pagination={{ pageSize: 20 }}
                  scroll={{ x: isSast ? 1650 : 1500 }}
                />
              ) : <Empty description={completed ? (isSast ? "当前规则未发现源码风险" : "未发现已收录漏洞") : "等待扫描完成"} />
            },
            ...(!isSast ? [{
              key: "components",
              label: `组件 ${scan.data?.componentCount ?? 0}`,
              children: components.data?.length ? (
                <Table
                  rowKey="id"
                  columns={componentColumns}
                  dataSource={components.data}
                  pagination={{ pageSize: 20 }}
                  scroll={{ x: 1400 }}
                />
              ) : <Empty description={completed ? "未解析到组件" : "等待扫描完成"} />
            }] : []),
            {
              key: "metadata",
              label: "扫描信息",
              children: scan.data ? (
                <Descriptions column={2} bordered size="small" className="scan-metadata">
                  <Descriptions.Item label="引擎版本">{scan.data.engineVersion || "—"}</Descriptions.Item>
                  <Descriptions.Item label="扫描类型">{scan.data.scanType}</Descriptions.Item>
                  <Descriptions.Item label="解析方式">{scan.data.sourceMode || "—"}</Descriptions.Item>
                  <Descriptions.Item label="创建时间">{formatTime(scan.data.createdAt)}</Descriptions.Item>
                  <Descriptions.Item label="完成时间">{formatTime(scan.data.finishedAt)}</Descriptions.Item>
                  <Descriptions.Item label="源码到期">{formatTime(scan.data.sourceExpiresAt)}</Descriptions.Item>
                  <Descriptions.Item label="结论到期">{formatTime(scan.data.resultExpiresAt)}</Descriptions.Item>
                </Descriptions>
              ) : null
            }
          ]}
        />
      </Card>
    </div>
  );
}

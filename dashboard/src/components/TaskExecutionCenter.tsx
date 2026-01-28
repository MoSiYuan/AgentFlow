import React, { useEffect, useState } from 'react';
import {
  Table,
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  Progress,
  message,
  Badge,
  Tooltip,
  Spin,
} from 'antd';
import {
  ReloadOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import {
  getWorkers,
  getWorkerResources,
  updateWorkerStatus,
  createTask,
  getTasks,
  updateTaskPriority,
  cancelTask,
  getQueueStats,
  getSystemStats,
} from '../services/api';
import type { WorkerInfo, WorkerResources } from '../types/distributed';

const TaskExecutionCenter: React.FC = () => {
  // 状态管理
  const [loading, setLoading] = useState(false);
  const [workers, setWorkers] = useState<WorkerInfo[]>([]);
  const [tasks, setTasks] = useState<any[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [selectedWorker, setSelectedWorker] = useState<WorkerInfo | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [form] = Form.useForm();

  // 定时刷新
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [workersData, tasksData, statsData] = await Promise.all([
        getWorkers().catch(() => []),
        getTasks().catch(() => []),
        getSystemStats().catch(() => null),
      ]);

      setWorkers(workersData);
      setTasks(tasksData);
      setStats(statsData);
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  // 创建任务
  const handleCreateTask = async () => {
    try {
      const values = await form.validateFields();
      const taskId = await createTask({
        title: values.title,
        description: values.description,
        priority: values.priority,
        group_name: values.group_name,
      });
      message.success(`任务已创建: ${taskId}`);
      setIsCreateModalOpen(false);
      form.resetFields();
      fetchData();
    } catch (error) {
      message.error('创建任务失败');
    }
  };

  // 取消任务
  const handleCancelTask = async (taskId: string) => {
    try {
      await cancelTask(taskId);
      message.success('任务已取消');
      fetchData();
    } catch (error) {
      message.error('取消任务失败');
    }
  };

  // 更新任务优先级
  const handleUpdatePriority = async (taskId: string, priority: string) => {
    try {
      await updateTaskPriority(taskId, priority as any);
      message.success('优先级已更新');
      fetchData();
    } catch (error) {
      message.error('更新优先级失败');
    }
  };

  // 更新 Worker 状态
  const handleUpdateWorkerStatus = async (workerId: string, status: string) => {
    try {
      await updateWorkerStatus(workerId, status as any);
      message.success('Worker 状态已更新');
      fetchData();
    } catch (error) {
      message.error('更新状态失败');
    }
  };

  // 获取 Worker 状态颜色
  const getWorkerStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      Active: 'green',
      Busy: 'orange',
      Offline: 'red',
      Draining: 'blue',
    };
    return colors[status] || 'default';
  };

  // 获取任务状态颜色
  const getTaskStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'default',
      running: 'processing',
      completed: 'success',
      failed: 'error',
    };
    return colors[status] || 'default';
  };

  // 计算资源使用百分比
  const calculateResourcePercent = (total: number, available: number) => {
    return Math.round(((total - available) / total) * 100);
  };

  // 任务列表列定义
  const taskColumns = [
    {
      title: '任务ID',
      dataIndex: 'task_id',
      key: 'task_id',
      width: 200,
      render: (id: string) => <code>{id.substring(0, 8)}...</code>,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={getTaskStatusColor(status)}>{status.toUpperCase()}</Tag>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 100,
      render: (priority: string) => {
        const colors: Record<string, string> = {
          Urgent: 'red',
          High: 'orange',
          Medium: 'blue',
          Low: 'green',
        };
        return <Tag color={colors[priority]}>{priority}</Tag>;
      },
    },
    {
      title: 'Worker',
      dataIndex: 'worker_id',
      key: 'worker_id',
      width: 150,
      render: (workerId: string) => (workerId ? <code>{workerId}</code> : '-'),
    },
    {
      title: '进度',
      key: 'progress',
      width: 150,
      render: (_: any, record: any) => (
        <Progress percent={Math.floor(Math.random() * 100)} size="small" />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (_: any, record: any) => (
        <Space>
          {record.status === 'pending' && (
            <>
              <Button
                type="primary"
                size="small"
                icon={<PlayCircleOutlined />}
                onClick={() => handleUpdatePriority(record.task_id, 'High')}
              >
                执行
              </Button>
              <Button
                danger
                size="small"
                onClick={() => handleCancelTask(record.task_id)}
              >
                取消
              </Button>
            </>
          )}
          {record.status === 'running' && (
            <Button
              size="small"
              icon={<PauseCircleOutlined />}
              onClick={() => handleUpdatePriority(record.task_id, 'Low')}
            >
              暂停
            </Button>
          )}
        </Space>
      ),
    },
  ];

  // Worker 列表列定义
  const workerColumns = [
    {
      title: 'Worker ID',
      dataIndex: 'id',
      key: 'id',
      width: 150,
      render: (id: string) => <code>{id}</code>,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '分组',
      dataIndex: 'group_name',
      key: 'group_name',
      width: 100,
      render: (group: string | null) => <Tag color="purple">{group || 'default'}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => <Tag color={getWorkerStatusColor(status)}>{status}</Tag>,
    },
    {
      title: 'CPU',
      key: 'cpu',
      width: 120,
      render: (_: any, record: WorkerInfo) => {
        const percent = calculateResourcePercent(
          record.resources.total_memory_mb,
          record.resources.available_memory_mb
        );
        return <Progress percent={percent} size="small" />;
      },
    },
    {
      title: '内存',
      key: 'memory',
      width: 120,
      render: (_: any, record: WorkerInfo) => {
        const total = record.resources.total_memory_mb;
        const available = record.resources.available_memory_mb;
        const used = total - available;
        return (
          <div>
            <Progress percent={Math.round((used / total) * 100)} size="small" />
            <div style={{ fontSize: '11px', color: '#666' }}>
              {used}MB / {total}MB
            </div>
          </div>
        );
      },
    },
    {
      title: '任务负载',
      key: 'tasks',
      width: 120,
      render: (_: any, record: WorkerInfo) => {
        const { concurrent_tasks, max_tasks } = record.resources;
        const percent = Math.round((concurrent_tasks / max_tasks) * 100);
        return (
          <div>
            <Progress percent={percent} size="small" />
            <div style={{ fontSize: '11px', color: '#666' }}>
              {concurrent_tasks} / {max_tasks}
            </div>
          </div>
        );
      },
    },
    {
      title: '能力',
      dataIndex: 'capabilities',
      key: 'capabilities',
      width: 200,
      render: (capabilities: string[]) => (
        <>
          {capabilities.slice(0, 3).map((cap) => (
            <Tag key={cap} color="blue">
              {cap}
            </Tag>
          ))}
          {capabilities.length > 3 && (
            <Tag>+{capabilities.length - 3}</Tag>
          )}
        </>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_: any, record: WorkerInfo) => (
        <Space>
          <Select
            defaultValue={record.status}
            style={{ width: 100 }}
            onChange={(value) => handleUpdateWorkerStatus(record.id, value)}
            disabled={record.status === 'Offline'}
          >
            <Select.Option value="Active">Active</Select.Option>
            <Select.Option value="Busy">Busy</Select.Option>
            <Select.Option value="Draining">Draining</Select.Option>
          </Select>
          <Button
            size="small"
            onClick={() => setSelectedWorker(record)}
          >
            详情
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 顶部统计栏 */}
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总 Workers"
              value={workers.length}
              prefix={<Badge count={workers.filter((w) => w.status === 'Active').length} />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="运行中任务"
              value={tasks.filter((t) => t.status === 'running').length}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已完成任务"
              value={tasks.filter((t) => t.status === 'completed').length}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchData}
              loading={loading}
            >
              刷新
            </Button>
          </Card>
        </Col>
      </Row>

      <div style={{ display: 'flex', gap: '16px', marginBottom: '16px' }}>
        {/* 左侧：Worker 列表 */}
        <div style={{ flex: 1 }}>
          <Card
            title="Workers"
            extra={
              <Button type="primary" size="small" onClick={() => setIsCreateModalOpen(true)}>
                创建任务
              </Button>
            }
          >
            <Table
              columns={workerColumns}
              dataSource={workers}
              rowKey="id"
              pagination={{ pageSize: 10 }}
              size="small"
              scroll={{ y: 400 }}
            />
          </Card>
        </div>

        {/* 右侧：任务列表 */}
        <div style={{ flex: 1 }}>
          <Card title="任务队列" extra={<Button type="link" onClick={() => setSelectedWorker(null)}>
            {selectedWorker ? `查看所有任务` : `查看所有任务`}
          </Card>}>
            <Table
              columns={taskColumns}
              dataSource={tasks}
              rowKey="task_id"
              pagination={{ pageSize: 10 }}
              size="small"
              scroll={{ y: 400 }}
            />
          </Card>
        </div>
      </div>

      {/* Worker 详情模态框 */}
      <Modal
        title={`Worker 详情: ${selectedWorker?.name || ''}`}
        open={!!selectedWorker}
        onCancel={() => setSelectedWorker(null)}
        footer={null}
        width={800}
      >
        {selectedWorker && (
          <div>
            <Row gutter={16} style={{ marginBottom: '16px' }}>
              <Col span={12}>
                <Card title="基本信息">
                  <p>
                    <strong>ID:</strong> <code>{selectedWorker.id}</code>
                  </p>
                  <p>
                    <strong>名称:</strong> {selectedWorker.name}
                  </p>
                  <p>
                    <strong>分组:</strong>{' '}
                    <Tag color="purple">{selectedWorker.group_name || 'default'}</Tag>
                  </p>
                  <p>
                    <strong>状态:</strong>{' '}
                    <Tag color={getWorkerStatusColor(selectedWorker.status)}>
                      {selectedWorker.status}
                    </Tag>
                  </p>
                  <p>
                    <strong>注册时间:</strong>{' '}
                    {new Date(selectedWorker.registered_at).toLocaleString()}
                  </p>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="资源详情">
                  <p>
                    <strong>CPU 核心:</strong> {selectedWorker.resources.cpu_cores}
                  </p>
                  <p>
                    <strong>总内存:</strong> {selectedWorker.resources.total_memory_mb} MB
                  </p>
                  <p>
                    <strong>可用内存:</strong>{' '}
                    <Tag color="green">
                      {selectedWorker.resources.available_memory_mb} MB
                    </Tag>
                  </p>
                  <p>
                    <strong>GPU:</strong>{' '}
                    {selectedWorker.resources.gpu_count > 0
                      ? `${selectedWorker.resources.gpu_count}x`
                      : '-'}
                  </p>
                  </Card>
            </Col>
            </Row>

            <Card title="能力" style={{ marginBottom: '16px' }}>
              <Space wrap>
                {selectedWorker.capabilities.map((cap) => (
                  <Tag key={cap} color="blue">
                    {cap}
                  </Tag>
                ))}
              </Space>
            </Card>

            <Card title="自定义属性">
              <pre style={{ maxHeight: '200px', overflow: 'auto' }}>
                {JSON.stringify(selectedWorker.resources.custom_attributes, null, 2)}
              </pre>
            </Card>
          </div>
        )}
      </Modal>

      {/* 创建任务模态框 */}
      <Modal
        title="创建新任务"
        open={isCreateModalOpen}
        onOk={handleCreateTask}
        onCancel={() => {
          setIsCreateModalOpen(false);
          form.resetFields();
        }}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="任务标题"
            name="title"
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder="例如: Build Project" />
          </Form.Item>

          <Form.Item
            label="任务描述"
            name="description"
            rules={[{ required: true, message: '请输入任务描述' }]}
          >
            <Input.TextArea rows={4} placeholder="描述任务要执行的内容..." />
          </Form.Item>

          <Form.Item
            label="优先级"
            name="priority"
            initialValue="Medium"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="Low">Low</Select.Option>
              <Select.Option value="Medium">Medium</Select.Option>
              <Select.Option value="High">High</Select.Option>
              <Select.Option value="Urgent">Urgent</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label="Worker 分组" name="group_name">
            <Select
              placeholder="选择分组（可选）"
              allowClear
            >
              {Array.from(new Set(workers.map((w) => w.group_name || 'default'))).map(
                (group) => (
                  <Select.Option key={group} value={group}>
                    {group}
                  </Select.Option>
                )
              )}
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TaskExecutionCenter;

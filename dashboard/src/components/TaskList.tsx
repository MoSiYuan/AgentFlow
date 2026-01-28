import React, { useEffect, useState } from 'react';
import { Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { TaskInfo } from '../types/messages';

const TaskList: React.FC = () => {
  const [tasks, setTasks] = useState<TaskInfo[]>([]);

  useEffect(() => {
    // 模拟数据
    setTasks([
      {
        task_id: 'task-1',
        project_id: 'diveadstra',
        status: 'running',
        progress: '60%',
        created_at: '2026-01-28T10:30:00Z',
      },
      {
        task_id: 'task-2',
        project_id: 'myproject',
        status: 'completed',
        progress: '100%',
        created_at: '2026-01-28T09:15:00Z',
      },
    ]);
  }, []);

  const columns: ColumnsType<TaskInfo> = [
    { title: 'Task ID', dataIndex: 'task_id', key: 'task_id' },
    { title: 'Project', dataIndex: 'project_id', key: 'project_id' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const color = status === 'running' ? 'blue' : status === 'completed' ? 'green' : 'red';
        return <Tag color={color}>{status.toUpperCase()}</Tag>;
      },
    },
    { title: 'Progress', dataIndex: 'progress', key: 'progress' },
    { title: 'Created At', dataIndex: 'created_at', key: 'created_at' },
  ];

  return <Table dataSource={tasks} columns={columns} rowKey="task_id" />;
};

export default TaskList;

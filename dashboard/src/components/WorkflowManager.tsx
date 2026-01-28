import React, { useEffect, useState } from 'react';
import { Card, Spin } from 'antd';

const WorkflowManager: React.FC = () => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => setLoading(false), 500);
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin tip="加载工作流管理..." />
      </div>
    );
  }

  return (
    <div style={{ padding: '20px', textAlign: 'center' }}>
      <Card>
        <h2>📊 工作流管理</h2>
        <p>工作流管理功能正在开发中...</p>
        <p style={{ color: '#666' }}>
          即将支持：DAG 工作流创建、任务依赖管理、执行顺序可视化
        </p>
      </Card>
    </div>
  );
};

export default WorkflowManager;

import React, { useState } from 'react';
import { Layout, Tabs } from 'antd';
import ClusterTopology from './ClusterTopology';
import TaskExecutionCenter from './TaskExecutionCenter';
import LockManager from './LockManager';
import LogStream from './LogStream';

const { Header, Content, Footer } = Layout;

const DashboardLayout: React.FC = () => {
  const [activeTab, setActiveTab] = useState('tasks');

  const tabItems = [
    {
      key: 'tasks',
      label: '任务执行中心',
      children: <TaskExecutionCenter />,
    },
    {
      key: 'cluster',
      label: '集群拓扑',
      children: <ClusterTopology />,
    },
    {
      key: 'locks',
      label: '分布式锁',
      children: <LockManager />,
    },
    {
      key: 'logs',
      label: '日志流',
      children: <LogStream />,
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ background: '#001529', padding: '0 24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1 style={{ color: '#fff', lineHeight: '64px', margin: 0 }}>
            AgentFlow 分布式执行系统
          </h1>
          <div style={{ color: '#fff' }}>
            v0.4.0
          </div>
        </div>
      </Header>

      <Content style={{ padding: '24px', background: '#f0f2f5' }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          type="card"
          items={tabItems}
          style={{ background: '#fff', padding: '16px', borderRadius: '4px' }}
        />
      </Content>

      <Footer style={{ background: '#000', padding: '16px 24px' }}>
        <LogStream />
      </Footer>
    </Layout>
  );
};

export default DashboardLayout;

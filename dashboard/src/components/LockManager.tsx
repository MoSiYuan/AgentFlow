import React, { useEffect, useState } from 'react';
import { Card, Spin } from 'antd';

const LockManager: React.FC = () => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => setLoading(false), 500);
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin tip="加载分布式锁管理..." />
      </div>
    );
  }

  return (
    <div style={{ padding: '20px', textAlign: 'center' }}>
      <Card>
        <h2>🔒 分布式锁管理</h2>
        <p>分布式锁管理功能正在开发中...</p>
        <p style={{ color: '#666' }}>
          即将支持：锁状态查看、获取新锁、释放锁、TTL 管理
        </p>
      </Card>
    </div>
  );
};

export default LockManager;

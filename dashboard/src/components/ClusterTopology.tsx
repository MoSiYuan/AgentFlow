import React, { useEffect, useState } from 'react';
import { Spin } from 'antd';

const ClusterTopology: React.FC = () => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 模拟加载
    setTimeout(() => setLoading(false), 1000);
  }, []);

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '500px'
      }}>
        <Spin size="large" tip="加载集群拓扑..." />
      </div>
    );
  }

  return (
    <div style={{ padding: '20px', textAlign: 'center' }}>
      <h2>🏛️ 集群拓扑</h2>
      <p>集群拓扑可视化正在加载中...</p>
      <p style={{ color: '#666' }}>
        功能包括：Master 节点、Worker 节点、实时连接关系
      </p>
      <div style={{
        marginTop: '20px',
        padding: '20px',
        border: '1px dashed #d9d9d9',
        borderRadius: '4px'
      }}>
        <strong>示例集群结构：</strong>
        <ul style={{ textAlign: 'left', marginTop: '10px' }}>
          <li>👑 Leader: master-1</li>
          <li>🏛️ Master: master-2, master-3</li>
          <li>⚙️ Workers: worker-1 (builders), worker-2 (testers)</li>
        </ul>
      </div>
    </div>
  );
};

export default ClusterTopology;

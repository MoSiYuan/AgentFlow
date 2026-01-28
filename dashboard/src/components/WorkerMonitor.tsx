import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Spin } from 'antd';

const WorkerMonitor: React.FC = () => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => setLoading(false), 500);
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin tip="加载 Worker 监控..." />
      </div>
    );
  }

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: '16px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="总 Workers" value={3} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="活跃" value={2} valueStyle={{ color: '#3f8600' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="繁忙" value={1} valueStyle={{ color: '#cf1322' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="离线" value={0} valueStyle={{ color: '#999' }} />
          </Card>
        </Col>
      </Row>

      <div style={{ padding: '20px', textAlign: 'center' }}>
        <p>Worker 监控功能正在开发中...</p>
        <p style={{ color: '#666' }}>
          即将支持：实时状态、资源监控、负载查看
        </p>
      </div>
    </div>
  );
};

export default WorkerMonitor;

import React, { useEffect, useState } from 'react';

interface LogEntry {
  level: string;
  content: string;
  timestamp: string;
}

const LogStream: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);

  useEffect(() => {
    // 模拟日志
    const mockLogs = [
      { level: 'info', content: '[Master 1] Task started: task-1', timestamp: '10:30:00' },
      { level: 'bash', content: '$ cargo build --release', timestamp: '10:30:05' },
      { level: 'info', content: '[Master 2] Task completed: task-2', timestamp: '10:30:10' },
    ];
    setLogs(mockLogs);
  }, []);

  const getLogColor = (level: string) => {
    switch (level) {
      case 'bash': return '#FFD700'; // 黄色
      case 'info': return '#FFFFFF'; // 白色
      case 'error': return '#FF6347'; // 红色
      default: return '#808080'; // 灰色
    }
  };

  return (
    <div style={{
      height: '100%',
      background: '#1E1E1E',
      color: '#fff',
      padding: '8px',
      overflowY: 'auto',
      fontFamily: 'monospace',
      fontSize: '12px',
    }}>
      {logs.map((log, index) => (
        <div key={index} style={{ color: getLogColor(log.level) }}>
          [{log.timestamp}] {log.content}
        </div>
      ))}
    </div>
  );
};

export default LogStream;

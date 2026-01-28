import React, { useState } from 'react';
import { Input, Select } from 'antd';
import { SendOutlined } from '@ant-design/icons';

const ChatPanel: React.FC = () => {
  const [messages, setMessages] = useState<Array<{role: string, text: string}>>([
    { role: 'assistant', text: '您好！我是 Claude，请问需要帮您做什么？' }
  ]);
  const [input, setInput] = useState('');
  const [targetContext, setTargetContext] = useState<string>('all:all');

  const handleSend = () => {
    if (!input.trim()) return;

    // 添加用户消息
    setMessages(prev => [...prev, { role: 'user', text: input }]);

    // TODO: 发送到 WebSocket
    console.log('Send to', targetContext, ':', input);

    setInput('');
  };

  return (
    <div style={{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Select
        value={targetContext}
        onChange={setTargetContext}
        style={{ marginBottom: '16px' }}
        options={[
          { value: 'all:all', label: '全部节点' },
          { value: 'diveadstra:master-1', label: 'DiveAdstra - Master 1' },
          { value: 'myproject:master-2', label: 'MyProject - Master 2' },
        ]}
      />

      <div style={{ flex: 1, overflowY: 'auto', marginBottom: '16px' }}>
        {messages.map((msg, index) => (
          <div key={index} style={{ marginBottom: '12px' }}>
            <strong>{msg.role}:</strong> {msg.text}
          </div>
        ))}
      </div>

      <Input.Search
        value={input}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setInput(e.target.value)}
        onSearch={handleSend}
        enterButton={<SendOutlined />}
        placeholder="输入指令..."
      />
    </div>
  );
};

export default ChatPanel;

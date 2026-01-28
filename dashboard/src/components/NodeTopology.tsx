import React, { useEffect, useState } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
} from 'reactflow';
import 'reactflow/dist/style.css';

const NodeTopology: React.FC = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);

  useEffect(() => {
    // 模拟数据
    const leaderNode: Node = {
      id: 'leader-1',
      type: 'default',
      data: { label: 'Leader (管理节点)' },
      position: { x: 400, y: 100 },
      style: { background: '#90EE90', border: '2px solid #000' },
    };

    const masterNodes: Node[] = [
      {
        id: 'master-1',
        type: 'default',
        data: { label: 'Master 1 (CPU: 45%)' },
        position: { x: 200, y: 300 },
        style: { background: '#87CEEB', border: '2px solid #000' },
      },
      {
        id: 'master-2',
        type: 'default',
        data: { label: 'Master 2 (CPU: 32%)' },
        position: { x: 600, y: 300 },
        style: { background: '#87CEEB', border: '2px solid #000' },
      },
    ];

    const edges: Edge[] = [
      { id: 'leader-master1', source: 'leader-1', target: 'master-1', animated: true },
      { id: 'leader-master2', source: 'leader-1', target: 'master-2', animated: true },
    ];

    setNodes([leaderNode, ...masterNodes]);
    setEdges(edges);
  }, []);

  return (
    <div style={{ height: '500px', width: '100%' }}>
      <ReactFlow nodes={nodes} edges={edges} fitView>
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  );
};

export default NodeTopology;

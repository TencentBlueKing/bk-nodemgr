
export const areaList = [
  {
    bk_networkarea_id: 0,
    bk_networkarea_name: '云电脑',
  },
  {
    bk_networkarea_id: 1,
    bk_networkarea_name: '默认区域',
  },
  {
    bk_networkarea_id: 2,
    bk_networkarea_name: '上海公有云',
  },
  {
    bk_networkarea_id: 3,
    bk_networkarea_name: 'test',
  },
  {
    bk_networkarea_id: 4,
    bk_networkarea_name: 'DBM 专用云区域',
  },
];

export const nodeList = [
  {
    id: 'node-1',
    data: {
      name: 'default',
      proxy: 1,
      agent: 200,
    },
    combo: 'combo1',
  },
  {
    id: 'node-2',
    data: {
      name: 'default',
      proxy: 2,
      agent: 200,
    },
    combo: 'combo3'
  },
  {
    id: 'node-11',
    data: {
      name: 'default',
      proxy: 11,
      agent: 500,
    },
    combo: 'combo2'
  },
  {
    id: 'node-12',
    data: {
      name: 'K8s',
      proxy: 12,
      agent: 800,
    },
    combo: 'combo2'
  },
  {
    id: 'node-13',
    data: {
      name: 'default2',
      proxy: 13,
      agent: 150,
    },
    combo: 'combo2'
  },
  {
    id: 'node-14',
    data: {
      name: 'Container',
      proxy: 14,
      agent: 400,
    },
    combo: 'combo2'
  },
  {
    id: 'node-3',
    data: {
      name: 'DevCloud',
      proxy: 3,
      agent: 200,
    },
    combo: 'combo3'
  },
  {
    id: 'node-4',
    data: {
      name: 'default',
      proxy: 4,
      agent: 1200,
    },
    combo: 'combo3'
  },
  {
    id: 'node-5',
    data: {
      name: 'default',
      proxy: 5,
      agent: 200,
    },
    combo: 'combo4'
  },
  {
    id: 'node-6',
    data: {
      name: 'CloudHost',
      proxy: 6,
      agent: 300,
    },
    combo: 'combo1'
  },
  {
    id: 'node-7',
    data: {
      name: 'default',
      proxy: 7,
      agent: 500,
    },
    combo: 'combo4'
  },
  {
    id: 'node-8',
    data: {
      name: 'K8s',
      proxy: 8,
      agent: 800,
    },
    combo: 'combo1'
  },
  {
    id: 'node-9',
    data: {
      name: 'default',
      proxy: 9,
      agent: 150,
    },
    combo: 'combo3'
  },
  {
    id: 'node-10',
    data: {
      name: 'Container',
      proxy: 10,
      agent: 400,
    },
    combo: 'combo4'
  },
  {
    id: 'node-15',
    data: {
      name: 'default',
      proxy: 15,
      agent: 500,
    },
    combo: 'combo5'
  },
  {
    id: 'node-16',
    data: {
      name: 'K8s',
      proxy: 16,
      agent: 800,
    },
    combo: 'combo5'
  },
  {
    id: 'node-17',
    data: {
      name: 'default1',
      proxy: 17,
      agent: 150,
    },
    combo: 'combo5'
  },
  {
    id: 'node-18',
    data: {
      name: 'Container',
      proxy: 18,
      agent: 400,
    },
    combo: 'combo5'
  },
];

export const edgeList = [
  { id: 'edge-1', source: 'node-1', target: 'node-2' },
  { id: 'edge-2', source: 'node-3', target: 'node-2' },
  { id: 'edge-3', source: 'node-4', target: 'node-2' },
  { id: 'edge-4', source: 'node-5', target: 'node-3' },
  { id: 'edge-5', source: 'node-6', target: 'node-1' },
  { id: 'edge-6', source: 'node-7', target: 'node-5' },
  { id: 'edge-7', source: 'node-8', target: 'node-6' },
  { id: 'edge-8', source: 'node-9', target: 'node-4' },
  { id: 'edge-9', source: 'node-10', target: 'node-7' },
  { id: 'edge-10', source: 'node-3', target: 'node-8' }
];
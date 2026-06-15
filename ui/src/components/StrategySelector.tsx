import { Radio, Tooltip, Space, Tag } from 'antd'
import type { ChunkStrategy } from '../types/chunk'

interface Props {
  value: ChunkStrategy
  onChange: (s: ChunkStrategy) => void
}

const strategies: { key: ChunkStrategy; label: string; desc: string; tag: string }[] = [
  {
    key: 'free',
    label: 'Free',
    desc: '按分隔符优先级递归分割，适合纯文本和日志文件',
    tag: '通用',
  },
  {
    key: 'md',
    label: 'Markdown',
    desc: '解析 Markdown AST，按标题层级切分，保留章节路径',
    tag: '推荐',
  },
  {
    key: 'eino',
    label: 'Semantic',
    desc: '基于 Embedding 向量余弦相似度在语义边界处切分，需配置 Embedding 服务',
    tag: 'AI',
  },
  {
    key: 'hierarchical',
    label: 'Hierarchical',
    desc: '两层父子结构：小子块精确检索 + 大父块提供上下文',
    tag: '高级',
  },
]

// StrategySelector 切块策略选择器，展示四种策略供用户选择。
export default function StrategySelector({ value, onChange }: Props) {
  return (
    <Radio.Group
      value={value}
      onChange={(e) => onChange(e.target.value as ChunkStrategy)}
      optionType="button"
      buttonStyle="solid"
      size="middle"
      style={{ width: '100%' }}
    >
      <Space direction="vertical" style={{ width: '100%' }} size="small">
        {strategies.map((s) => (
          <Tooltip key={s.key} title={s.desc} placement="right">
            <Radio.Button
              value={s.key}
              style={{
                width: '100%',
                height: 'auto',
                padding: '6px 12px',
                textAlign: 'left',
              }}
            >
              <Space>
                <span style={{ fontWeight: 600 }}>{s.label}</span>
                <Tag
                  color={value === s.key ? 'blue' : 'default'}
                  style={{ fontSize: 10, lineHeight: '16px', padding: '0 4px' }}
                >
                  {s.tag}
                </Tag>
              </Space>
            </Radio.Button>
          </Tooltip>
        ))}
      </Space>
    </Radio.Group>
  )
}

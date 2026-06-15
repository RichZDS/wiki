import { Slider, InputNumber, Select, Space, Row, Col, Button } from 'antd'
import { UndoOutlined } from '@ant-design/icons'
import type { ChunkStrategy } from '../types/chunk'

interface Props {
  chunkSize: number
  chunkOverlap: number
  separators: string[]
  strategy: ChunkStrategy
  onChunkSizeChange: (v: number) => void
  onChunkOverlapChange: (v: number) => void
  onSeparatorsChange: (v: string[]) => void
}

const SEPARATOR_SUGGESTIONS = ['\\n\\n', '\\n', '。', '.', '，', ',']

const PRESETS: { label: string; size: number; overlap: number }[] = [
  { label: '默认 (500/50)', size: 500, overlap: 50 },
  { label: '细粒度 (200/20)', size: 200, overlap: 20 },
  { label: '粗粒度 (1000/100)', size: 1000, overlap: 100 },
]

// ParamPanel 切块参数调节面板，含 ChunkSize、ChunkOverlap、Separators 及预设。
export default function ParamPanel({
  chunkSize,
  chunkOverlap,
  separators,
  strategy,
  onChunkSizeChange,
  onChunkOverlapChange,
  onSeparatorsChange,
}: Props) {
  const showSeparators = strategy === 'free'

  const applyPreset = (size: number, overlap: number) => {
    onChunkSizeChange(size)
    onChunkOverlapChange(overlap)
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="small">
      <Row gutter={8}>
        <Col span={12}>
          <div style={{ marginBottom: 4, fontSize: 13, color: '#666' }}>Chunk Size</div>
          <Slider
            min={50}
            max={2000}
            step={50}
            value={chunkSize}
            onChange={onChunkSizeChange}
          />
        </Col>
        <Col span={12}>
          <InputNumber
            min={50}
            max={2000}
            step={50}
            value={chunkSize}
            onChange={(v) => v && onChunkSizeChange(v)}
            style={{ width: '100%' }}
            addonAfter="字符"
          />
        </Col>
      </Row>

      <Row gutter={8}>
        <Col span={12}>
          <div style={{ marginBottom: 4, fontSize: 13, color: '#666' }}>Chunk Overlap</div>
          <Slider
            min={0}
            max={500}
            step={10}
            value={chunkOverlap}
            onChange={onChunkOverlapChange}
          />
        </Col>
        <Col span={12}>
          <InputNumber
            min={0}
            max={500}
            step={10}
            value={chunkOverlap}
            onChange={(v) => v !== null && onChunkOverlapChange(v)}
            style={{ width: '100%' }}
            addonAfter="字符"
          />
        </Col>
      </Row>

      {showSeparators && (
        <div>
          <div style={{ marginBottom: 4, fontSize: 13, color: '#666' }}>Separators</div>
          <Select
            mode="tags"
            style={{ width: '100%' }}
            placeholder="输入分隔符，如 \\n\\n, \\n, 。"
            value={separators}
            onChange={onSeparatorsChange}
            options={SEPARATOR_SUGGESTIONS.map((s) => ({ label: s, value: s }))}
          />
        </div>
      )}

      <div>
        <div style={{ marginBottom: 4, fontSize: 13, color: '#666' }}>预设</div>
        <Space wrap>
          {PRESETS.map((p) => (
            <Button
              key={p.label}
              size="small"
              type={chunkSize === p.size && chunkOverlap === p.overlap ? 'primary' : 'default'}
              onClick={() => applyPreset(p.size, p.overlap)}
            >
              {p.label}
            </Button>
          ))}
          <Button size="small" icon={<UndoOutlined />} onClick={() => applyPreset(500, 50)}>
            重置
          </Button>
        </Space>
      </div>
    </Space>
  )
}

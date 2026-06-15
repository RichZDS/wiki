import { useState, useCallback } from 'react'
import {
  Row, Col, Tabs, message, Spin, Card, Space, Button,
} from 'antd'
import {
  ScissorOutlined,
  SwapOutlined,
  ApartmentOutlined,
} from '@ant-design/icons'
import TextInput from '../components/TextInput'
import StrategySelector from '../components/StrategySelector'
import ParamPanel from '../components/ParamPanel'
import ResultSection from '../components/ResultSection'
import CompareView from '../components/CompareView'
import HierarchicalTree from '../components/HierarchicalTree'
import { chunk, compare } from '../api/chunk'
import type {
  ChunkStrategy,
  ChunkResult,
  ChunkCompareResult,
} from '../types/chunk'

const DEFAULT_STRATEGY: ChunkStrategy = 'md'

export default function Playground() {
  // 输入状态
  const [content, setContent] = useState('')
  const [strategy, setStrategy] = useState<ChunkStrategy>(DEFAULT_STRATEGY)
  const [chunkSize, setChunkSize] = useState(500)
  const [chunkOverlap, setChunkOverlap] = useState(50)
  const [separators, setSeparators] = useState<string[]>([])

  // 结果状态
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ChunkResult | null>(null)
  const [compareResult, setCompareResult] = useState<ChunkCompareResult | null>(null)
  const [activeTab, setActiveTab] = useState('single')

  // 执行单策略切块
  const handleChunk = useCallback(async () => {
    if (!content.trim()) {
      message.warning('请输入要切分的文本')
      return
    }
    setLoading(true)
    try {
      const res = await chunk({
        content,
        strategy,
        chunk_size: chunkSize,
        chunk_overlap: chunkOverlap,
        separators,
      })
      setResult(res)
      setCompareResult(null)
      setActiveTab('single')
      message.success(`切块完成：${res.stats.total_chunks} 个块`)
    } catch (e) {
      message.error('切块失败：' + (e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [content, strategy, chunkSize, chunkOverlap, separators])

  // 执行四策略对比
  const handleCompare = useCallback(async () => {
    if (!content.trim()) {
      message.warning('请输入要切分的文本')
      return
    }
    setLoading(true)
    try {
      const res = await compare({
        content,
        configs: [
          { strategy: 'free', chunk_size: chunkSize, chunk_overlap: chunkOverlap, separators },
          { strategy: 'md', chunk_size: chunkSize, chunk_overlap: chunkOverlap, separators: [] },
          { strategy: 'eino', chunk_size: chunkSize, chunk_overlap: chunkOverlap, separators: [] },
          { strategy: 'hierarchical', chunk_size: chunkSize, chunk_overlap: chunkOverlap, separators: [] },
        ],
      })
      setCompareResult(res)
      setResult(null)
      setActiveTab('compare')
      const successCount = res.results.length
      message.success(`对比完成：${successCount} 个策略成功`)
    } catch (e) {
      message.error('对比失败：' + (e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [content, chunkSize, chunkOverlap, separators])

  return (
    <Spin spinning={loading} tip="切块中...">
      <Row gutter={24}>
        {/* 左侧：输入 + 配置 */}
        <Col xs={24} lg={7}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Card title="📝 文本输入" size="small">
              <TextInput value={content} onChange={setContent} />
            </Card>

            <Card title="⚙️ 切块策略" size="small">
              <StrategySelector value={strategy} onChange={setStrategy} />
            </Card>

            <Card title="🎛️ 参数调节" size="small">
              <ParamPanel
                chunkSize={chunkSize}
                chunkOverlap={chunkOverlap}
                separators={separators}
                strategy={strategy}
                onChunkSizeChange={setChunkSize}
                onChunkOverlapChange={setChunkOverlap}
                onSeparatorsChange={setSeparators}
              />
            </Card>

            <Space>
              <Button
                type="primary"
                icon={<ScissorOutlined />}
                onClick={handleChunk}
                size="large"
              >
                执行切块
              </Button>
              <Button
                icon={<SwapOutlined />}
                onClick={handleCompare}
                size="large"
              >
                四种策略对比
              </Button>
            </Space>
          </Space>
        </Col>

        {/* 右侧：结果展示 */}
        <Col xs={24} lg={17}>
          <Card size="small">
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={[
                {
                  key: 'single',
                  label: '📋 单策略结果',
                  children: result ? (
                    <ResultSection result={result} />
                  ) : (
                    <div style={{ color: '#999', textAlign: 'center', padding: 60 }}>
                      输入文本并点击「执行切块」查看结果
                    </div>
                  ),
                },
                {
                  key: 'compare',
                  label: '🔄 策略对比',
                  children: compareResult ? (
                    <CompareView result={compareResult} />
                  ) : (
                    <div style={{ color: '#999', textAlign: 'center', padding: 60 }}>
                      点击「四种策略对比」查看各策略效果
                    </div>
                  ),
                },
                {
                  key: 'tree',
                  label: <span><ApartmentOutlined /> 分层视图</span>,
                  children: compareResult ? (
                    <HierarchicalTree compareResult={compareResult} />
                  ) : result ? (
                    <HierarchicalTree singleResult={result} />
                  ) : (
                    <div style={{ color: '#999', textAlign: 'center', padding: 60 }}>
                      使用分层策略（Hierarchical）切块后查看父子关系
                    </div>
                  ),
                },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </Spin>
  )
}

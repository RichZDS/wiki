import { useState, useCallback } from 'react'
import {
  Row, Col, Tabs, message, Spin, Card, Space, Button,
  Input, Radio, Slider, InputNumber, Tag, Collapse, Empty, Typography,
} from 'antd'
import {
  CloudUploadOutlined,
  SearchOutlined,
  FileTextOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { ingest, search } from '../api/rag'
import type {
  ChunkStrategy,
  RAGIngestResult,
  RAGSearchResult,
  RAGSearchItem,
} from '../types/rag'

const { TextArea } = Input
const { Text, Paragraph } = Typography

const DOC_SAMPLES: Record<string, string> = {
  eino: `Eino 是 CloudWeGo 开源的 AI 应用开发框架，基于 Golang 构建。
Eino 提供了丰富的组件抽象，包括 ChatModel、Embedding、Indexer、Retriever 等。
通过 Eino，开发者可以快速搭建 RAG（检索增强生成）应用。
Eino 的 Chunk 模块支持多种文档切块策略：
1. Free Chunk - 基于字符分隔符的递归切分
2. Markdown Chunk - 基于 Markdown AST 的结构化切分
3. Eino Semantic Chunk - 基于向量相似度的语义切分
4. Hierarchical Chunk - 父子层级上下文切分

Eino 的 Embedding 模块支持 OpenAI、Ark（火山引擎）、Gemini 等多种模型。
Eino 的 Indexer 和 Retriever 组件支持 Redis、Elasticsearch 等向量存储后端。
整个框架遵循组件化设计，各个模块可独立使用或组合成流水线。`,
  redis: `Redis 是一个开源的内存数据结构存储系统，可用作数据库、缓存和消息代理。
Redis 支持多种数据结构，包括字符串、哈希、列表、集合、有序集合等。
在 AI 应用中，Redis 常用于向量存储和检索。
Redis Stack 提供了 RediSearch 模块，支持向量相似度搜索（FT.SEARCH）。
使用 Redis 作为向量数据库的优势：
1. 极低延迟 - 内存存储，查询速度极快
2. 丰富的查询语法 - 支持 KNN、范围搜索、过滤器
3. 高可用 - 支持主从复制和集群模式
4. 持久化 - 支持 RDB 和 AOF 两种持久化方式

RediSearch 支持两种向量搜索模式：
- KNN 向量搜索：返回 Top-K 最相似的文档
- 向量范围搜索：返回距离阈值内的所有文档`,
}

export default function RagPlayground() {
  // 入库状态
  const [content, setContent] = useState('')
  const [strategy, setStrategy] = useState<ChunkStrategy>('free')
  const [chunkSize, setChunkSize] = useState(500)
  const [chunkOverlap, setChunkOverlap] = useState(50)
  const [docPrefix, setDocPrefix] = useState('wiki')

  // 检索状态
  const [query, setQuery] = useState('')

  // 结果状态
  const [loading, setLoading] = useState(false)
  const [ingestResult, setIngestResult] = useState<RAGIngestResult | null>(null)
  const [searchResult, setSearchResult] = useState<RAGSearchResult | null>(null)
  const [activeTab, setActiveTab] = useState('ingest')

  const handleIngest = useCallback(async () => {
    if (!content.trim()) {
      message.warning('请输入要入库的文本')
      return
    }
    setLoading(true)
    try {
      const res = await ingest({
        content,
        strategy,
        chunk_size: chunkSize,
        chunk_overlap: chunkOverlap,
        doc_id_prefix: docPrefix,
      })
      setIngestResult(res)
      setActiveTab('ingest')
      message.success(`入库完成：${res.chunk_count} 个文档块`)
    } catch (e) {
      message.error('入库失败：' + (e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [content, strategy, chunkSize, chunkOverlap, docPrefix])

  const handleSearch = useCallback(async () => {
    if (!query.trim()) {
      message.warning('请输入检索关键词')
      return
    }
    setLoading(true)
    try {
      const res = await search({ query: query.trim() })
      setSearchResult(res)
      setActiveTab('search')
      if (res.results.length === 0) {
        message.info('未找到匹配的文档')
      } else {
        message.success(`找到 ${res.results.length} 个相关文档`)
      }
    } catch (e) {
      message.error('检索失败：' + (e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [query])

  const handleLoadSample = useCallback((key: string) => {
    setContent(DOC_SAMPLES[key])
  }, [])

  return (
    <Spin spinning={loading} tip="处理中...">
      <Row gutter={24}>
        {/* 左侧操作区 */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>

            {/* 入库区域 */}
            <Card
              title={<span><CloudUploadOutlined /> 文档入库</span>}
              size="small"
            >
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <TextArea
                  value={content}
                  onChange={e => setContent(e.target.value)}
                  placeholder="粘贴或输入要入库的文档内容..."
                  rows={8}
                  style={{ fontFamily: 'monospace', fontSize: 13 }}
                />

                <Space size="small" wrap>
                  <Button size="small" onClick={() => handleLoadSample('eino')}>
                    加载 Eino 文档
                  </Button>
                  <Button size="small" onClick={() => handleLoadSample('redis')}>
                    加载 Redis 文档
                  </Button>
                  <Button
                    size="small"
                    icon={<DeleteOutlined />}
                    danger
                    onClick={() => setContent('')}
                    disabled={!content}
                  >
                    清空
                  </Button>
                </Space>

                <Text type="secondary" style={{ fontSize: 12 }}>切块策略</Text>
                <Radio.Group
                  value={strategy}
                  onChange={e => setStrategy(e.target.value)}
                  optionType="button"
                  buttonStyle="solid"
                  size="small"
                >
                  <Radio.Button value="free">Free</Radio.Button>
                  <Radio.Button value="md">MD</Radio.Button>
                  <Radio.Button value="eino">Semantic</Radio.Button>
                  <Radio.Button value="hierarchical">Hier</Radio.Button>
                </Radio.Group>

                <Collapse size="small" ghost items={[{
                  key: 'params',
                  label: '切块参数设置',
                  children: (
                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                      <Row align="middle" justify="space-between">
                        <Text style={{ fontSize: 12, width: 80 }}>块大小</Text>
                        <Slider
                          min={100} max={2000} step={50}
                          value={chunkSize}
                          onChange={setChunkSize}
                          style={{ flex: 1, margin: '0 8px' }}
                        />
                        <InputNumber
                          min={100} max={2000}
                          value={chunkSize}
                          onChange={v => setChunkSize(v ?? 500)}
                          size="small"
                          style={{ width: 70 }}
                        />
                      </Row>
                      <Row align="middle" justify="space-between">
                        <Text style={{ fontSize: 12, width: 80 }}>重叠大小</Text>
                        <Slider
                          min={0} max={500} step={10}
                          value={chunkOverlap}
                          onChange={setChunkOverlap}
                          style={{ flex: 1, margin: '0 8px' }}
                        />
                        <InputNumber
                          min={0} max={500}
                          value={chunkOverlap}
                          onChange={v => setChunkOverlap(v ?? 50)}
                          size="small"
                          style={{ width: 70 }}
                        />
                      </Row>
                      <Row align="middle" justify="space-between">
                        <Text style={{ fontSize: 12, width: 80 }}>文档前缀</Text>
                        <Input
                          value={docPrefix}
                          onChange={e => setDocPrefix(e.target.value)}
                          size="small"
                          style={{ flex: 1 }}
                          placeholder="wiki"
                        />
                      </Row>
                    </Space>
                  ),
                }]} />

                <Button
                  type="primary"
                  icon={<CloudUploadOutlined />}
                  onClick={handleIngest}
                  block
                  size="large"
                >
                  执行入库
                </Button>
              </Space>
            </Card>

            {/* 检索区域 */}
            <Card
              title={<span><SearchOutlined /> 语义检索</span>}
              size="small"
            >
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <Input
                  value={query}
                  onChange={e => setQuery(e.target.value)}
                  placeholder="输入检索问题或关键词..."
                  onPressEnter={handleSearch}
                  prefix={<SearchOutlined style={{ color: '#bbb' }} />}
                  allowClear
                />
                <Button
                  icon={<SearchOutlined />}
                  onClick={handleSearch}
                  block
                  size="large"
                >
                  执行检索
                </Button>
              </Space>
            </Card>

          </Space>
        </Col>

        {/* 右侧结果区 */}
        <Col xs={24} lg={16}>
          <Card size="small" style={{ minHeight: 500 }}>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={[
                {
                  key: 'ingest',
                  label: <span><CloudUploadOutlined /> 入库结果</span>,
                  children: ingestResult ? (
                    <IngestResultView result={ingestResult} />
                  ) : (
                    <EmptyPlaceholder text="粘贴文档并点击「执行入库」" />
                  ),
                },
                {
                  key: 'search',
                  label: <span><SearchOutlined /> 检索结果</span>,
                  children: searchResult ? (
                    <SearchResultView result={searchResult} query={query} />
                  ) : (
                    <EmptyPlaceholder text="输入关键词并点击「执行检索」" />
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

function EmptyPlaceholder({ text }: { text: string }) {
  return (
    <div style={{ color: '#999', textAlign: 'center', padding: 80 }}>
      <FileTextOutlined style={{ fontSize: 48, marginBottom: 16 }} />
      <div>{text}</div>
    </div>
  )
}

function IngestResultView({ result }: { result: RAGIngestResult }) {
  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Row gutter={16}>
        <Col span={8}>
          <Card size="small" style={{ textAlign: 'center', background: '#f6ffed' }}>
            <div style={{ fontSize: 24, fontWeight: 700, color: '#52c41a' }}>
              {result.chunk_count}
            </div>
            <Text type="secondary">切块数量</Text>
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small" style={{ textAlign: 'center', background: '#e6f7ff' }}>
            <div style={{ fontSize: 24, fontWeight: 700, color: '#1677ff' }}>
              {result.total_chars.toLocaleString()}
            </div>
            <Text type="secondary">总字符数</Text>
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small" style={{ textAlign: 'center', background: '#fff7e6' }}>
            <div style={{ fontSize: 24, fontWeight: 700, color: '#fa8c16' }}>
              {result.stored_ids.length}
            </div>
            <Text type="secondary">已存储 ID</Text>
          </Card>
        </Col>
      </Row>

      <Card size="small" title="存储的文档 ID 列表">
        <div style={{ maxHeight: 300, overflow: 'auto' }}>
          {result.stored_ids.map((id, i) => (
            <Tag key={id} color="blue" style={{ marginBottom: 4 }}>
              {i + 1}. {id}
            </Tag>
          ))}
        </div>
      </Card>
    </Space>
  )
}

function SearchResultView({ result, query }: { result: RAGSearchResult; query: string }) {
  if (result.results.length === 0) {
    return (
      <Empty
        description={`未找到与 "${query}" 相关的文档`}
        style={{ padding: 60 }}
      />
    )
  }

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Text type="secondary">
        共找到 <Text strong>{result.results.length}</Text> 个相关文档
      </Text>

      {result.results.map((item, i) => (
        <SearchResultCard key={item.id} item={item} index={i} />
      ))}
    </Space>
  )
}

function SearchResultCard({ item, index }: { item: RAGSearchItem; index: number }) {
  const scoreColor = item.score != null
    ? item.score > 0.7 ? '#52c41a' : item.score > 0.4 ? '#fa8c16' : '#ff4d4f'
    : '#999'

  return (
    <Card
      size="small"
      title={
        <Space>
          <Tag color="blue">#{index + 1}</Tag>
          <Text ellipsis style={{ maxWidth: 480 }}>
            {item.id}
          </Text>
        </Space>
      }
      extra={
        item.score != null ? (
          <Tag color={item.score > 0.7 ? 'success' : item.score > 0.4 ? 'warning' : 'error'}>
            相似度 {(item.score * 100).toFixed(1)}%
          </Tag>
        ) : null
      }
      style={{ borderLeft: `3px solid ${scoreColor}` }}
    >
      <Paragraph
        ellipsis={{ rows: 3, expandable: true, symbol: '展开' }}
        style={{ marginBottom: 8, whiteSpace: 'pre-wrap' }}
      >
        {item.content}
      </Paragraph>

      {item.metadata && Object.keys(item.metadata).length > 0 && (
        <Collapse
          size="small"
          ghost
          items={[{
            key: 'meta',
            label: <Text type="secondary" style={{ fontSize: 11 }}>元数据</Text>,
            children: (
              <div style={{ fontSize: 12, maxHeight: 200, overflow: 'auto' }}>
                <pre style={{ margin: 0 }}>
                  {JSON.stringify(item.metadata, null, 2)}
                </pre>
              </div>
            ),
          }]}
        />
      )}
    </Card>
  )
}

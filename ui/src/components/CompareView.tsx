import { Row, Col, Card, Statistic, Alert, Typography, Collapse } from 'antd'
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import type { ChunkCompareResult } from '../types/chunk'
import ChunkCard from './ChunkCard'

const { Text } = Typography

const STRATEGY_NAMES: Record<string, string> = {
  free: 'Free 自由切块',
  md: 'Markdown 切块',
  eino: 'Semantic 语义切块',
  hierarchical: 'Hierarchical 分层切块',
}

const STRATEGY_COLORS: Record<string, string> = {
  free: '#52c41a',
  md: '#1677ff',
  eino: '#722ed1',
  hierarchical: '#fa8c16',
}

interface Props {
  result: ChunkCompareResult
}

// CompareView 四策略对比视图，2x2 网格并排展示各策略结果。
export default function CompareView({ result }: Props) {
  const { results, errors } = result

  return (
    <div>
      {/* 错误提示 */}
      {errors && Object.keys(errors).length > 0 && (
        <div style={{ marginBottom: 16 }}>
          {Object.entries(errors).map(([strategy, msg]) => (
            <Alert
              key={strategy}
              type="warning"
              showIcon
              icon={<ExclamationCircleOutlined />}
              message={`${STRATEGY_NAMES[strategy] || strategy}: ${msg}`}
              description={
                strategy === 'eino'
                  ? '请设置 ARK_API_KEY 或 OPENAI_API_KEY 环境变量启用语义切块'
                  : undefined
              }
              style={{ marginBottom: 8 }}
            />
          ))}
        </div>
      )}

      {/* 2x2 网格 */}
      <Row gutter={[16, 16]}>
        {['free', 'md', 'eino', 'hierarchical'].map((strategy) => {
          const found = results.find((r) => r.strategy === strategy)
          const hasError = errors?.[strategy]

          return (
            <Col xs={24} md={12} key={strategy}>
              <Card
                size="small"
                title={
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <div
                      style={{
                        width: 12,
                        height: 12,
                        borderRadius: '50%',
                        backgroundColor: STRATEGY_COLORS[strategy],
                      }}
                    />
                    <Text strong>{STRATEGY_NAMES[strategy]}</Text>
                    {found && <CheckCircleOutlined style={{ color: '#52c41a' }} />}
                    {hasError && <CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
                  </div>
                }
                style={{
                  borderColor: found
                    ? STRATEGY_COLORS[strategy]
                    : hasError
                      ? '#ff4d4f'
                      : undefined,
                  borderTopWidth: 2,
                }}
              >
                {found ? (
                  <div>
                    <Row gutter={[8, 4]} style={{ marginBottom: 12 }}>
                      <Col span={12}>
                        <Statistic title="块数" value={found.stats.total_chunks} valueStyle={{ fontSize: 18 }} />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="平均长度"
                          value={found.stats.avg_chunk_length}
                          precision={0}
                          valueStyle={{ fontSize: 18 }}
                        />
                      </Col>
                    </Row>
                    <div style={{ maxHeight: 400, overflow: 'auto' }}>
                      {found.chunks.map((item, idx) => (
                        <ChunkCard key={item.id} item={item} index={idx} />
                      ))}
                    </div>
                  </div>
                ) : hasError ? (
                  <div style={{ textAlign: 'center', padding: 20, color: '#999' }}>
                    不可用
                  </div>
                ) : null}
              </Card>
            </Col>
          )
        })}
      </Row>
    </div>
  )
}

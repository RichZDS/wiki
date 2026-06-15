import { Row, Col, Statistic, Collapse, Space, Typography, Empty } from 'antd'
import { FileTextOutlined, BarChartOutlined } from '@ant-design/icons'
import type { ChunkResult } from '../types/chunk'
import ChunkCard from './ChunkCard'

const { Text } = Typography

interface Props {
  result: ChunkResult
}

// ResultSection 单策略切块结果展示，包含统计信息和块列表。
export default function ResultSection({ result }: Props) {
  const { stats, chunks } = result

  if (chunks.length === 0) {
    return <Empty description="无切块结果" />
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="middle">
      {/* 统计栏 */}
      <Row gutter={[16, 8]}>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="总块数"
            value={stats.total_chunks}
            prefix={<FileTextOutlined />}
            valueStyle={{ color: '#1677ff' }}
          />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="总字符数"
            value={stats.total_characters}
            valueStyle={{ color: '#52c41a' }}
          />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="平均长度"
            value={stats.avg_chunk_length}
            precision={0}
            prefix={<BarChartOutlined />}
            valueStyle={{ color: '#722ed1' }}
          />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="最小块"
            value={stats.min_chunk_length}
            valueStyle={{ color: '#faad14' }}
          />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="最大块"
            value={stats.max_chunk_length}
            valueStyle={{ color: '#ff4d4f' }}
          />
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Statistic
            title="策略"
            value={
              chunks[0]?.metadata?.chunk_strategy
                ? String(chunks[0].metadata.chunk_strategy)
                : '-'
            }
          />
        </Col>
      </Row>

      {/* 块列表 */}
      <div>
        <Text type="secondary" style={{ marginBottom: 8, display: 'block' }}>
          共 {chunks.length} 个块
        </Text>
        {chunks.map((item, idx) => (
          <ChunkCard key={item.id} item={item} index={idx} />
        ))}
      </div>
    </Space>
  )
}

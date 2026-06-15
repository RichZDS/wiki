import { Card, Typography, Tag, Space, Button, Tooltip, Badge } from 'antd'
import { CopyOutlined, FileTextOutlined } from '@ant-design/icons'
import type { ChunkItem } from '../types/chunk'

const { Paragraph, Text } = Typography

interface Props {
  item: ChunkItem
  index: number
}

// ChunkCard 单个切块卡片，展示内容预览、元数据和操作按钮。
export default function ChunkCard({ item, index }: Props) {
  const headingPath = item.metadata?.heading_path as string | undefined
  const chunkRole = item.metadata?.chunk_role as string | undefined
  const chunkStrategy = item.metadata?.chunk_strategy as string | undefined
  const elementTypes = item.metadata?.element_types as string[] | undefined
  const childIDs = item.metadata?.child_chunk_ids as string[] | undefined

  const handleCopy = () => {
    navigator.clipboard.writeText(item.content).then(() => {
      // Let antd message handle notification at the parent level
    })
  }

  const strategyColorMap: Record<string, string> = {
    free: 'green',
    md: 'blue',
    eino: 'purple',
    hierarchical: 'orange',
  }

  return (
    <Card
      size="small"
      title={
        <Space size="small">
          <Badge count={index + 1} style={{ backgroundColor: '#1677ff' }} />
          {chunkRole && (
            <Tag color={chunkRole === 'parent' ? 'orange' : 'cyan'}>
              {chunkRole}
            </Tag>
          )}
          {chunkStrategy && (
            <Tag color={strategyColorMap[chunkStrategy] || 'default'}>
              {chunkStrategy}
            </Tag>
          )}
          <Text type="secondary" style={{ fontSize: 12 }}>
            {item.length} 字符
          </Text>
        </Space>
      }
      extra={
        <Tooltip title="复制内容">
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={handleCopy}
          />
        </Tooltip>
      }
      style={{ marginBottom: 8 }}
    >
      <Space direction="vertical" style={{ width: '100%' }} size="small">
        {headingPath && (
          <div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              📂 {headingPath}
            </Text>
          </div>
        )}

        {childIDs && childIDs.length > 0 && (
          <div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              子块: {childIDs.join(', ')}
            </Text>
          </div>
        )}

        <Paragraph
          ellipsis={{ rows: 4, expandable: true, symbol: '展开' }}
          style={{
            marginBottom: 0,
            padding: '8px 12px',
            background: '#fafafa',
            borderRadius: 6,
            fontFamily: 'monospace',
            fontSize: 13,
            whiteSpace: 'pre-wrap',
          }}
        >
          {item.content}
        </Paragraph>

        {elementTypes && elementTypes.length > 0 && (
          <Space size={4} wrap>
            {elementTypes.map((t) => (
              <Tag key={t} style={{ fontSize: 11 }}>{t}</Tag>
            ))}
          </Space>
        )}
      </Space>
    </Card>
  )
}

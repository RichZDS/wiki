import React, { useState, useMemo } from 'react'
import { Tree, Drawer, Typography, Tag, Empty } from 'antd'
import type { ChunkCompareResult, ChunkResult, ChunkItem } from '../types/chunk'

const { Paragraph, Text } = Typography

interface DataNode {
  title: React.ReactNode
  key: string
  children?: DataNode[]
  chunk?: ChunkItem
}

interface Props {
  compareResult?: ChunkCompareResult
  singleResult?: ChunkResult
}

// HierarchicalTree 分层切块树形视图，展示父子块层级关系。
export default function HierarchicalTree({ compareResult, singleResult }: Props) {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedChunk, setSelectedChunk] = useState<ChunkItem | null>(null)

  // 从结果中提取分层策略的 chunks
  const hierarchicalChunks = useMemo(() => {
    if (compareResult) {
      const hr = compareResult.results.find((r) => r.strategy === 'hierarchical')
      return hr?.chunks || []
    }
    if (singleResult) {
      return singleResult.chunks
    }
    return []
  }, [compareResult, singleResult])

  // 构建树形数据
  const treeData = useMemo(() => {
    const parentMap = new Map<string, ChunkItem>()
    const childMap = new Map<string, ChunkItem[]>()

    for (const chunk of hierarchicalChunks) {
      const role = chunk.metadata?.chunk_role as string | undefined
      if (role === 'parent') {
        parentMap.set(chunk.id, chunk)
      } else if (role === 'child') {
        const parentID = chunk.metadata?.parent_chunk_id as string | undefined
        if (parentID) {
          const siblings = childMap.get(parentID) || []
          siblings.push(chunk)
          childMap.set(parentID, siblings)
        }
      }
    }

    // 如果不是分层结果，退化为扁平列表
    if (parentMap.size === 0) {
      return hierarchicalChunks.map((chunk, idx) => ({
        title: `${chunk.id} (${chunk.length} 字符)` as React.ReactNode,
        key: chunk.id || `chunk-${idx}`,
        chunk,
      }))
    }

    const nodes: DataNode[] = []
    for (const [parentID, parent] of parentMap) {
      const children = childMap.get(parentID) || []
      const childIDs = parent.metadata?.child_chunk_ids as string[] | undefined

      const childCount = childIDs ? childIDs.length : 0
      nodes.push({
        title: `🏷️ [parent] ${parentID} (${parent.length} 字符${childCount > 0 ? `, ${childCount} 个子块` : ''})`,
        key: parentID,
        chunk: parent,
        children: children.map((child) => ({
          title: `  ↳ [child] ${child.id} (${child.length} 字符)`,
          key: child.id,
          chunk: child,
        })),
      })
    }
    return nodes
  }, [hierarchicalChunks])

  const onSelect = (_keys: React.Key[], info: { node: DataNode }) => {
    if (info.node.chunk) {
      setSelectedChunk(info.node.chunk)
      setDrawerOpen(true)
    }
  }

  if (hierarchicalChunks.length === 0) {
    return (
      <Empty
        description={
          <span>
            暂无分层切块数据。请使用 <Tag color="orange">Hierarchical 分层切块</Tag>策略生成。
          </span>
        }
      />
    )
  }

  return (
    <div>
      <Tree
        treeData={treeData}
        defaultExpandAll
        showLine
        onSelect={onSelect as never}
        style={{ padding: '8px 0' }}
      />

      <Drawer
        title="块详情"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={520}
      >
        {selectedChunk && (
          <div>
            <div style={{ marginBottom: 16 }}>
              <Text strong>ID: </Text>
              <Tag>{selectedChunk.id}</Tag>
              <Tag>{selectedChunk.length} 字符</Tag>
            </div>

            <div style={{ marginBottom: 12 }}>
              <Text strong>角色: </Text>
              <Tag color={
                (selectedChunk.metadata?.chunk_role as string) === 'parent'
                  ? 'orange'
                  : 'cyan'
              }>
                {String(selectedChunk.metadata?.chunk_role || '-')}
              </Tag>
            </div>

            {(selectedChunk.metadata?.heading_path as string | undefined) && (
              <div style={{ marginBottom: 12 }}>
                <Text strong>标题路径: </Text>
                <Text>{String(selectedChunk.metadata.heading_path)}</Text>
              </div>
            )}

            {(selectedChunk.metadata?.parent_content as string | undefined) && (
              <div style={{ marginBottom: 12 }}>
                <Text strong>父块内容:</Text>
                <Paragraph
                  ellipsis={{ rows: 4, expandable: true }}
                  style={{
                    padding: 8,
                    background: '#fafafa',
                    borderRadius: 6,
                    fontSize: 12,
                  }}
                >
                  {String(selectedChunk.metadata.parent_content)}
                </Paragraph>
              </div>
            )}

            <div style={{ marginBottom: 12 }}>
              <Text strong>内容:</Text>
              <Paragraph
                copyable
                style={{
                  padding: 12,
                  background: '#fafafa',
                  borderRadius: 6,
                  fontFamily: 'monospace',
                  fontSize: 13,
                  whiteSpace: 'pre-wrap',
                }}
              >
                {selectedChunk.content}
              </Paragraph>
            </div>

            <div>
              <Text strong>元数据:</Text>
              <pre style={{
                padding: 12,
                background: '#f5f5f5',
                borderRadius: 6,
                fontSize: 12,
                overflow: 'auto',
              }}>
                {JSON.stringify(selectedChunk.metadata, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </Drawer>
    </div>
  )
}

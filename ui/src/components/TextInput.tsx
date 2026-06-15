import { Input, Button, Space } from 'antd'
import { ClearOutlined, FileTextOutlined } from '@ant-design/icons'

const { TextArea } = Input

interface Props {
  value: string
  onChange: (v: string) => void
}

// TextInput 文本输入组件，支持加载示例文本。
export default function TextInput({ value, onChange }: Props) {
  const loadSampleMD = () => {
    onChange(`# Wiki 知识库建设指南

## 项目背景

Wiki 是一个面向团队的知识管理平台，支持文档的智能切块、向量化存储以及语义检索。
系统基于 RAG（检索增强生成）架构，将长文档切分为适合 LLM 处理的语义块。

## 核心功能

### 1. 文档切块引擎

文档切块是 RAG 流水线的核心环节，系统提供四种切块策略以适应不同场景：

- **自由切块（Free）**：按分隔符优先级递归分割，适合纯文本和日志文件
- **Markdown 切块（MD）**：解析 Markdown AST，按标题层级切分，保留章节路径
- **语义切块（Eino）**：基于 Embedding 向量的余弦相似度在语义边界处切分
- **分层切块（Hierarchical）**：两层父子结构，小子块用于精确检索，大父块提供上下文

### 2. 技术架构

系统采用 Go + Gin 作为后端框架，前端使用 React + Ant Design。
数据库采用 MySQL + GORM，缓存使用 Redis，AI 组件基于 CloudWeGo Eino 框架。

### 3. 配置说明

切块参数包括：
- ChunkSize（块大小）：控制每块的最大字符数，默认 500
- ChunkOverlap（块重叠）：相邻块之间的重叠字符数，默认 50
- Separators（分隔符）：仅自由切块生效，指定分割优先级列表

## 总结

选择合适的切块策略和参数对于 RAG 系统的检索质量至关重要。
建议根据实际文档类型进行试验对比，找到最佳配置。`)
  }

  const loadSamplePlain = () => {
    onChange(`本项目旨在构建一个企业级知识管理平台。

平台核心能力包括文档的智能解析、语义切块和向量检索。

用户上传文档后，系统自动进行文本提取和清洗。接着根据文档类型选择合适的切块策略：
对于 Markdown 格式的技术文档，推荐使用 Markdown 切块策略，它能保留标题层级关系；
对于纯文本日志或聊天记录，自由切块策略更为合适，它按分隔符优先级递归分割；
对于需要精准语义边界的场景，语义切块通过计算句子向量的余弦相似度来检测话题转折点；
对于需要同时兼顾检索精度和上下文完整性的场景，分层切块提供了父子两层的块结构。

切块完成后，系统将每个文本块通过 Embedding 模型转换为稠密向量存储在向量数据库中。
用户查询时，系统将查询文本同样转为向量，通过相似度搜索找到最相关的文本块，
然后由大语言模型基于这些检索结果生成最终回答。`)
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size="small">
      <Space wrap>
        <Button size="small" icon={<FileTextOutlined />} onClick={loadSampleMD}>
          加载 Markdown 示例
        </Button>
        <Button size="small" icon={<FileTextOutlined />} onClick={loadSamplePlain}>
          加载纯文本示例
        </Button>
        <Button
          size="small"
          icon={<ClearOutlined />}
          disabled={!value}
          onClick={() => onChange('')}
        >
          清空
        </Button>
      </Space>
      <TextArea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="在此粘贴或输入要切分的文本内容..."
        rows={18}
        style={{ fontFamily: 'monospace' }}
      />
      <div style={{ color: '#999', fontSize: 12, textAlign: 'right' }}>
        {value.length > 0 ? `${value.length} 字符` : ''}
      </div>
    </Space>
  )
}

import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { BrowserRouter, NavLink, Navigate, Route, Routes } from 'react-router-dom'
import clsx from 'clsx'
import { Browser as KawaiiBrowser, Planet } from 'react-kawaii'
import {
  Activity,
  ArrowRight,
  Bot,
  Clock3,
  Database,
  FileText,
  Pause,
  Play,
  RefreshCw,
  RotateCw,
  Scissors,
  Search,
  Settings2,
  Sparkles,
  Trash2,
  UploadCloud,
} from 'lucide-react'
import { chunk, compare } from './api/chunk'
import { ingest, search } from './api/rag'
import { listJobLogs, listJobs, runJobNow, startJob, stopJob } from './api/job'
import type { ChunkCompareResult, ChunkResult, ChunkStrategy } from './types/chunk'
import type { RAGIngestResult, RAGSearchResult } from './types/rag'
import type { JobLog, JobLogListResult, JobSnapshot } from './types/job'

const STRATEGIES: { value: ChunkStrategy; label: string }[] = [
  { value: 'free', label: 'Free' },
  { value: 'md', label: 'Markdown' },
  { value: 'eino', label: 'Semantic' },
  { value: 'hierarchical', label: 'Hierarchical' },
]

const SAMPLE_TEXT = `# RAG 系统说明

系统会先把文档切分成适合检索的块，再通过 embedding 写入 Redis 向量索引。
检索时，用户问题会被向量化，并在 Redis 中执行 TopK 相似度查询。

好的切块策略需要平衡上下文完整性和命中精度：
- Markdown 文档适合结构化切块
- 日志和纯文本适合自由切块
- 长知识库适合分层切块
- 语义边界敏感内容可以尝试 Semantic`

function AppLayout() {
  return (
    <div className="min-h-screen bg-paper text-ink">
      <header className="border-b border-slate-200 bg-white/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-4 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-ink text-white">
              <Bot size={22} />
            </div>
            <div>
              <div className="text-lg font-semibold">Wiki Console</div>
              <div className="text-sm text-slate-500">RAG, chunking, jobs</div>
            </div>
          </div>
          <nav className="flex flex-wrap gap-2">
            <NavItem to="/" icon={<Scissors size={18} />} label="切块" />
            <NavItem to="/rag" icon={<Search size={18} />} label="RAG" />
            <NavItem to="/jobs" icon={<Activity size={18} />} label="Job" />
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-6">
        <Routes>
          <Route path="/" element={<ChunkLab />} />
          <Route path="/rag" element={<RagConsole />} />
          <Route path="/jobs" element={<JobConsole />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  )
}

function NavItem({ to, icon, label }: { to: string; icon: ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        clsx(
          'inline-flex h-10 items-center gap-2 rounded-lg px-4 text-sm font-medium transition',
          isActive
            ? 'bg-ink text-white shadow-soft'
            : 'border border-slate-200 bg-white text-slate-600 hover:border-violet hover:text-violet',
        )
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}

function ChunkLab() {
  const [content, setContent] = useState(SAMPLE_TEXT)
  const [strategy, setStrategy] = useState<ChunkStrategy>('md')
  const [chunkSize, setChunkSize] = useState(500)
  const [chunkOverlap, setChunkOverlap] = useState(50)
  const [separators, setSeparators] = useState('\\n\\n,\\n,。,.')
  const [result, setResult] = useState<ChunkResult | null>(null)
  const [compareResult, setCompareResult] = useState<ChunkCompareResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [notice, setNotice] = useState('')

  const separatorList = useMemo(
    () => separators.split(',').map(item => item.trim()).filter(Boolean),
    [separators],
  )

  const handleChunk = useCallback(async () => {
    if (!content.trim()) {
      setNotice('请输入文本')
      return
    }
    setLoading(true)
    setNotice('')
    try {
      const data = await chunk({
        content,
        strategy,
        chunk_size: chunkSize,
        chunk_overlap: chunkOverlap,
        separators: strategy === 'free' ? separatorList : [],
      })
      setResult(data)
      setCompareResult(null)
      setNotice(`切块完成：${data.stats.total_chunks} 个块`)
    } catch (error) {
      setNotice((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [chunkOverlap, chunkSize, content, separatorList, strategy])

  const handleCompare = useCallback(async () => {
    if (!content.trim()) {
      setNotice('请输入文本')
      return
    }
    setLoading(true)
    setNotice('')
    try {
      const data = await compare({
        content,
        configs: STRATEGIES.map(item => ({
          strategy: item.value,
          chunk_size: chunkSize,
          chunk_overlap: chunkOverlap,
          separators: item.value === 'free' ? separatorList : [],
        })),
      })
      setCompareResult(data)
      setResult(null)
      setNotice(`对比完成：${data.results.length} 个策略成功`)
    } catch (error) {
      setNotice((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [chunkOverlap, chunkSize, content, separatorList])

  return (
    <Workspace
      title="切块工作台"
      icon={<Scissors size={22} />}
      aside={<KawaiiBrowser size={112} mood="happy" color="#59d6b3" />}
      notice={notice}
    >
      <div className="grid gap-5 lg:grid-cols-[380px_minmax(0,1fr)]">
        <section className="space-y-4">
          <Panel title="输入" icon={<FileText size={18} />}>
            <textarea
              className="control min-h-64 resize-y"
              value={content}
              onChange={event => setContent(event.target.value)}
            />
            <div className="mt-3 flex gap-2">
              <Button icon={<Sparkles size={17} />} onClick={() => setContent(SAMPLE_TEXT)}>
                示例
              </Button>
              <Button icon={<Trash2 size={17} />} variant="ghost" onClick={() => setContent('')}>
                清空
              </Button>
            </div>
          </Panel>

          <Panel title="参数" icon={<Settings2 size={18} />}>
            <Segmented value={strategy} onChange={setStrategy} />
            <NumberSlider label="块大小" value={chunkSize} min={100} max={2000} step={50} onChange={setChunkSize} />
            <NumberSlider label="重叠" value={chunkOverlap} min={0} max={500} step={10} onChange={setChunkOverlap} />
            {strategy === 'free' && (
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-slate-600">分隔符</span>
                <input className="control" value={separators} onChange={event => setSeparators(event.target.value)} />
              </label>
            )}
            <div className="grid grid-cols-2 gap-2 pt-1">
              <Button loading={loading} icon={<Scissors size={17} />} onClick={handleChunk}>
                执行切块
              </Button>
              <Button loading={loading} icon={<RotateCw size={17} />} variant="secondary" onClick={handleCompare}>
                策略对比
              </Button>
            </div>
          </Panel>
        </section>

        <section className="space-y-4">
          {result ? <ChunkResultView result={result} /> : null}
          {compareResult ? <CompareResultView result={compareResult} /> : null}
          {!result && !compareResult ? (
            <EmptyState title="等待切块结果" description="左侧参数会直接影响块数量、平均长度和上下文保留度。" />
          ) : null}
        </section>
      </div>
    </Workspace>
  )
}

function RagConsole() {
  const [content, setContent] = useState(SAMPLE_TEXT)
  const [strategy, setStrategy] = useState<ChunkStrategy>('md')
  const [chunkSize, setChunkSize] = useState(500)
  const [chunkOverlap, setChunkOverlap] = useState(50)
  const [docPrefix, setDocPrefix] = useState('wiki')
  const [query, setQuery] = useState('RAG 检索如何工作？')
  const [topK, setTopK] = useState(5)
  const [ingestResult, setIngestResult] = useState<RAGIngestResult | null>(null)
  const [searchResult, setSearchResult] = useState<RAGSearchResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [notice, setNotice] = useState('')

  const handleIngest = useCallback(async () => {
    if (!content.trim()) {
      setNotice('请输入要入库的文本')
      return
    }
    setLoading(true)
    setNotice('')
    try {
      const data = await ingest({
        content,
        strategy,
        chunk_size: chunkSize,
        chunk_overlap: chunkOverlap,
        doc_id_prefix: docPrefix,
      })
      setIngestResult(data)
      setNotice(`入库完成：${data.chunk_count} 个文档块`)
    } catch (error) {
      setNotice((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [chunkOverlap, chunkSize, content, docPrefix, strategy])

  const handleSearch = useCallback(async () => {
    if (!query.trim()) {
      setNotice('请输入检索内容')
      return
    }
    setLoading(true)
    setNotice('')
    try {
      const data = await search({ query: query.trim(), top_k: topK })
      setSearchResult(data)
      setNotice(`检索完成：${data.results.length} 条结果`)
    } catch (error) {
      setNotice((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [query, topK])

  return (
    <Workspace title="RAG 控制台" icon={<Search size={22} />} aside={<Planet size={116} mood="blissful" color="#8e7cff" />} notice={notice}>
      <div className="grid gap-5 xl:grid-cols-[420px_minmax(0,1fr)]">
        <section className="space-y-4">
          <Panel title="入库" icon={<UploadCloud size={18} />}>
            <textarea className="control min-h-56 resize-y" value={content} onChange={event => setContent(event.target.value)} />
            <div className="mt-3 grid grid-cols-2 gap-2">
              <label>
                <span className="mb-1 block text-sm font-medium text-slate-600">切块策略</span>
                <select className="control" value={strategy} onChange={event => setStrategy(event.target.value as ChunkStrategy)}>
                  {STRATEGIES.map(item => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span className="mb-1 block text-sm font-medium text-slate-600">文档前缀</span>
                <input className="control" value={docPrefix} onChange={event => setDocPrefix(event.target.value)} />
              </label>
            </div>
            <NumberSlider label="块大小" value={chunkSize} min={100} max={2000} step={50} onChange={setChunkSize} />
            <NumberSlider label="重叠" value={chunkOverlap} min={0} max={500} step={10} onChange={setChunkOverlap} />
            <Button loading={loading} icon={<Database size={17} />} onClick={handleIngest}>
              写入向量库
            </Button>
          </Panel>

          <Panel title="检索" icon={<Search size={18} />}>
            <input className="control" value={query} onChange={event => setQuery(event.target.value)} onKeyDown={event => event.key === 'Enter' && handleSearch()} />
            <NumberSlider label="TopK" value={topK} min={1} max={50} step={1} onChange={setTopK} />
            <Button loading={loading} icon={<ArrowRight size={17} />} variant="secondary" onClick={handleSearch}>
              执行检索
            </Button>
          </Panel>
        </section>

        <section className="space-y-4">
          {ingestResult ? (
            <div className="grid gap-3 md:grid-cols-3">
              <Stat label="块数量" value={ingestResult.chunk_count} />
              <Stat label="字符数" value={ingestResult.total_chars.toLocaleString()} />
              <Stat label="存储 ID" value={ingestResult.stored_ids.length} />
            </div>
          ) : null}
          {searchResult ? <SearchResultView result={searchResult} /> : null}
          {!ingestResult && !searchResult ? (
            <EmptyState title="等待 RAG 操作" description="可以先写入示例文本，再用 TopK 控制检索结果数量。" />
          ) : null}
        </section>
      </div>
    </Workspace>
  )
}

function JobConsole() {
  const [jobs, setJobs] = useState<JobSnapshot[]>([])
  const [selectedName, setSelectedName] = useState('')
  const [logs, setLogs] = useState<JobLogListResult | null>(null)
  const [level, setLevel] = useState('')
  const [loading, setLoading] = useState(false)
  const [notice, setNotice] = useState('')

  const selected = jobs.find(item => item.name === selectedName) ?? jobs[0]

  const refreshJobs = useCallback(async () => {
    setLoading(true)
    setNotice('')
    try {
      const data = await listJobs()
      setJobs(data)
      if (!selectedName && data.length > 0) {
        setSelectedName(data[0].name)
      }
    } catch (error) {
      setNotice((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [selectedName])

  const refreshLogs = useCallback(async () => {
    const name = selected?.name
    if (!name) {
      setLogs(null)
      return
    }
    try {
      const data = await listJobLogs(name, level, 1, 30)
      setLogs(data)
    } catch (error) {
      setNotice((error as Error).message)
    }
  }, [level, selected?.name])

  useEffect(() => {
    refreshJobs()
  }, [refreshJobs])

  useEffect(() => {
    refreshLogs()
  }, [refreshLogs])

  const mutateJob = useCallback(
    async (action: 'start' | 'stop' | 'run') => {
      if (!selected?.name) {
        return
      }
      setLoading(true)
      setNotice('')
      try {
        if (action === 'start') {
          await startJob(selected.name)
          setNotice('任务已启用')
        } else if (action === 'stop') {
          await stopJob(selected.name)
          setNotice('任务已停用')
        } else {
          await runJobNow(selected.name)
          setNotice('任务已触发')
        }
        await refreshJobs()
        await refreshLogs()
      } catch (error) {
        setNotice((error as Error).message)
      } finally {
        setLoading(false)
      }
    },
    [refreshJobs, refreshLogs, selected?.name],
  )

  return (
    <Workspace title="Job 控制台" icon={<Activity size={22} />} aside={<KawaiiBrowser size={112} mood="excited" color="#ff8f70" />} notice={notice}>
      <div className="grid gap-5 lg:grid-cols-[360px_minmax(0,1fr)]">
        <section className="panel overflow-hidden">
          <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
            <div className="font-semibold">任务列表</div>
            <button className="icon-button" onClick={refreshJobs} disabled={loading} title="刷新任务">
              <RefreshCw size={17} />
            </button>
          </div>
          <div className="divide-y divide-slate-100">
            {jobs.map(job => (
              <button
                key={job.name}
                type="button"
                className={clsx(
                  'flex w-full items-center justify-between px-4 py-3 text-left transition hover:bg-slate-50',
                  selected?.name === job.name && 'bg-violet/10',
                )}
                onClick={() => setSelectedName(job.name)}
              >
                <div>
                  <div className="font-medium">{job.name}</div>
                  <div className="text-xs text-slate-500">{formatInterval(job.interval_ms)}</div>
                </div>
                <StatusPill job={job} />
              </button>
            ))}
          </div>
        </section>

        <section className="space-y-4">
          {selected ? (
            <>
              <Panel title={selected.name} icon={<Clock3 size={18} />}>
                <div className="grid gap-3 md:grid-cols-4">
                  <Stat label="调度" value={selected.enabled ? '启用' : '停用'} />
                  <Stat label="运行" value={selected.running ? '运行中' : '空闲'} />
                  <Stat label="状态" value={selected.last_status || '-'} />
                  <Stat label="下次执行" value={selected.next_run ? formatDate(selected.next_run) : '-'} />
                </div>
                {selected.last_error ? <div className="mt-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{selected.last_error}</div> : null}
                <div className="mt-4 flex flex-wrap gap-2">
                  <Button icon={<Play size={17} />} onClick={() => mutateJob('start')} disabled={loading || selected.enabled}>
                    开始
                  </Button>
                  <Button icon={<Pause size={17} />} variant="danger" onClick={() => mutateJob('stop')} disabled={loading || !selected.enabled}>
                    关闭
                  </Button>
                  <Button icon={<RotateCw size={17} />} variant="secondary" onClick={() => mutateJob('run')} disabled={loading || selected.running}>
                    立即运行
                  </Button>
                </div>
              </Panel>

              <Panel title="日志" icon={<FileText size={18} />}>
                <div className="mb-3 flex flex-wrap items-center gap-2">
                  <select className="control max-w-40" value={level} onChange={event => setLevel(event.target.value)}>
                    <option value="">全部级别</option>
                    <option value="debug">debug</option>
                    <option value="info">info</option>
                    <option value="warn">warn</option>
                    <option value="error">error</option>
                  </select>
                  <Button icon={<RefreshCw size={17} />} variant="ghost" onClick={refreshLogs}>
                    刷新日志
                  </Button>
                </div>
                <LogTable logs={logs?.list ?? []} />
              </Panel>
            </>
          ) : (
            <EmptyState title="没有注册任务" description="服务启动后会在这里显示已注册的周期任务。" />
          )}
        </section>
      </div>
    </Workspace>
  )
}

function Workspace({
  title,
  icon,
  aside,
  notice,
  children,
}: {
  title: string
  icon: ReactNode
  aside: ReactNode
  notice: string
  children: ReactNode
}) {
  return (
    <div className="space-y-5">
      <section className="panel overflow-hidden">
        <div className="flex flex-col gap-4 p-5 md:flex-row md:items-center md:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-violet/15 text-violet">{icon}</div>
            <h1 className="text-xl font-semibold">{title}</h1>
          </div>
          <div className="flex items-center gap-4">
            {notice ? <div className="rounded-lg bg-slate-100 px-3 py-2 text-sm text-slate-700">{notice}</div> : null}
            <div className="hidden md:block">{aside}</div>
          </div>
        </div>
      </section>
      {children}
    </div>
  )
}

function Panel({ title, icon, children }: { title: string; icon: ReactNode; children: ReactNode }) {
  return (
    <section className="panel p-4">
      <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700">
        {icon}
        {title}
      </div>
      {children}
    </section>
  )
}

function Button({
  children,
  icon,
  variant = 'primary',
  loading = false,
  disabled = false,
  onClick,
}: {
  children: ReactNode
  icon?: ReactNode
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  loading?: boolean
  disabled?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      disabled={disabled || loading}
      onClick={onClick}
      className={clsx(
        'inline-flex h-10 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50',
        variant === 'primary' && 'bg-ink text-white hover:bg-slate-700',
        variant === 'secondary' && 'bg-violet text-white hover:bg-violet/90',
        variant === 'ghost' && 'border border-slate-200 bg-white text-slate-700 hover:border-violet hover:text-violet',
        variant === 'danger' && 'bg-coral text-white hover:bg-coral/90',
      )}
    >
      {loading ? <RefreshCw className="animate-spin" size={16} /> : icon}
      {children}
    </button>
  )
}

function Segmented({ value, onChange }: { value: ChunkStrategy; onChange: (value: ChunkStrategy) => void }) {
  return (
    <div className="grid grid-cols-2 gap-2 md:grid-cols-4">
      {STRATEGIES.map(item => (
        <button
          key={item.value}
          type="button"
          onClick={() => onChange(item.value)}
          className={clsx(
            'h-10 rounded-lg border px-3 text-sm font-medium transition',
            value === item.value ? 'border-ink bg-ink text-white' : 'border-slate-200 bg-white text-slate-600 hover:border-violet',
          )}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}

function NumberSlider({
  label,
  value,
  min,
  max,
  step,
  onChange,
}: {
  label: string
  value: number
  min: number
  max: number
  step: number
  onChange: (value: number) => void
}) {
  return (
    <div className="mt-3">
      <div className="mb-1 flex items-center justify-between text-sm">
        <span className="font-medium text-slate-600">{label}</span>
        <input
          className="h-8 w-24 rounded-lg border border-slate-200 px-2 text-right text-sm outline-none focus:border-violet focus:ring-2 focus:ring-violet/20"
          type="number"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={event => onChange(Number(event.target.value))}
        />
      </div>
      <input className="w-full accent-violet" type="range" min={min} max={max} step={step} value={value} onChange={event => onChange(Number(event.target.value))} />
    </div>
  )
}

function ChunkResultView({ result }: { result: ChunkResult }) {
  return (
    <>
      <div className="grid gap-3 md:grid-cols-5">
        <Stat label="块数" value={result.stats.total_chunks} />
        <Stat label="字符数" value={result.stats.total_characters.toLocaleString()} />
        <Stat label="平均" value={Math.round(result.stats.avg_chunk_length)} />
        <Stat label="最小" value={result.stats.min_chunk_length} />
        <Stat label="最大" value={result.stats.max_chunk_length} />
      </div>
      <div className="space-y-3">
        {result.chunks.map((item, index) => (
          <article key={item.id} className="panel p-4">
            <div className="mb-2 flex items-center justify-between gap-3">
              <div className="text-sm font-semibold">#{index + 1} {item.id}</div>
              <span className="rounded-lg bg-mint/15 px-2 py-1 text-xs font-medium text-emerald-700">{item.length} 字</span>
            </div>
            <p className="whitespace-pre-wrap text-sm leading-6 text-slate-700">{item.content}</p>
          </article>
        ))}
      </div>
    </>
  )
}

function CompareResultView({ result }: { result: ChunkCompareResult }) {
  return (
    <div className="grid gap-4 xl:grid-cols-2">
      {result.results.map(item => (
        <section key={item.strategy} className="panel p-4">
          <div className="mb-3 flex items-center justify-between">
            <div className="font-semibold">{item.strategy}</div>
            <span className="rounded-lg bg-violet/10 px-2 py-1 text-xs font-medium text-violet">{item.stats.total_chunks} 块</span>
          </div>
          <div className="mb-3 grid grid-cols-3 gap-2">
            <Stat label="平均" value={Math.round(item.stats.avg_chunk_length)} compact />
            <Stat label="最小" value={item.stats.min_chunk_length} compact />
            <Stat label="最大" value={item.stats.max_chunk_length} compact />
          </div>
          <div className="max-h-80 space-y-2 overflow-auto pr-1">
            {item.chunks.slice(0, 6).map((chunkItem, index) => (
              <div key={chunkItem.id} className="rounded-lg border border-slate-200 bg-slate-50 p-3 text-sm text-slate-700">
                <div className="mb-1 font-medium">#{index + 1}</div>
                <p className="line-clamp-3 whitespace-pre-wrap">{chunkItem.content}</p>
              </div>
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}

function SearchResultView({ result }: { result: RAGSearchResult }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-3 md:grid-cols-3">
        <Stat label="结果" value={result.results.length} />
        <Stat label="TopK" value={result.top_k} />
        <Stat label="耗时" value={`${result.duration_ms} ms`} />
      </div>
      {result.results.length === 0 ? (
        <EmptyState title="没有检索结果" description="可以先确认文档已入库，或调大 TopK 后重试。" />
      ) : (
        result.results.map((item, index) => (
          <article key={item.id} className="panel p-4">
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
              <div className="font-semibold">#{index + 1} {item.id}</div>
              {item.score != null ? <span className="rounded-lg bg-mint/15 px-2 py-1 text-xs font-medium text-emerald-700">相似度 {(item.score * 100).toFixed(1)}%</span> : null}
            </div>
            <p className="whitespace-pre-wrap text-sm leading-6 text-slate-700">{item.content}</p>
            {Object.keys(item.metadata ?? {}).length > 0 ? (
              <pre className="mt-3 max-h-40 overflow-auto rounded-lg bg-slate-950 p-3 text-xs text-slate-100">{JSON.stringify(item.metadata, null, 2)}</pre>
            ) : null}
          </article>
        ))
      )}
    </div>
  )
}

function Stat({ label, value, compact = false }: { label: string; value: ReactNode; compact?: boolean }) {
  return (
    <div className={clsx('rounded-lg border border-slate-200 bg-white', compact ? 'p-2' : 'p-3')}>
      <div className="text-xs font-medium text-slate-500">{label}</div>
      <div className={clsx('font-semibold text-ink', compact ? 'text-base' : 'text-xl')}>{value}</div>
    </div>
  )
}

function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="panel flex min-h-72 flex-col items-center justify-center p-8 text-center">
      <Planet size={104} mood="happy" color="#59d6b3" />
      <div className="mt-3 text-lg font-semibold">{title}</div>
      <p className="mt-1 max-w-md text-sm leading-6 text-slate-500">{description}</p>
    </div>
  )
}

function StatusPill({ job }: { job: JobSnapshot }) {
  const tone = job.running
    ? 'bg-violet/10 text-violet'
    : job.enabled
      ? 'bg-mint/15 text-emerald-700'
      : 'bg-slate-100 text-slate-500'
  return <span className={clsx('rounded-lg px-2 py-1 text-xs font-medium', tone)}>{job.running ? '运行中' : job.enabled ? '启用' : '停用'}</span>
}

function LogTable({ logs }: { logs: JobLog[] }) {
  if (logs.length === 0) {
    return <div className="rounded-lg border border-dashed border-slate-300 p-8 text-center text-sm text-slate-500">暂无日志</div>
  }
  return (
    <div className="overflow-hidden rounded-lg border border-slate-200">
      <div className="max-h-[520px] overflow-auto">
        <table className="w-full min-w-[680px] border-collapse text-left text-sm">
          <thead className="sticky top-0 bg-slate-50 text-xs uppercase text-slate-500">
            <tr>
              <th className="px-3 py-2">时间</th>
              <th className="px-3 py-2">级别</th>
              <th className="px-3 py-2">Run ID</th>
              <th className="px-3 py-2">内容</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {logs.map(item => (
              <tr key={item.id} className="bg-white">
                <td className="px-3 py-2 text-slate-500">{formatDate(item.created_at)}</td>
                <td className="px-3 py-2">
                  <span className={clsx('rounded-lg px-2 py-1 text-xs font-medium', logLevelTone(item.level))}>{item.level}</span>
                </td>
                <td className="px-3 py-2 font-mono text-xs text-slate-500">{item.run_id || '-'}</td>
                <td className="px-3 py-2 text-slate-700">{item.message}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function logLevelTone(level: string) {
  if (level === 'error') return 'bg-rose-100 text-rose-700'
  if (level === 'warn') return 'bg-amber-100 text-amber-700'
  if (level === 'debug') return 'bg-slate-100 text-slate-600'
  return 'bg-mint/15 text-emerald-700'
}

function formatInterval(ms: number) {
  const seconds = Math.round(ms / 1000)
  if (seconds < 60) return `${seconds} 秒`
  const minutes = Math.round(seconds / 60)
  if (minutes < 60) return `${minutes} 分钟`
  return `${Math.round(minutes / 60)} 小时`
}

function formatDate(value: string) {
  return new Date(value).toLocaleString()
}

export default function App() {
  return (
    <BrowserRouter>
      <AppLayout />
    </BrowserRouter>
  )
}

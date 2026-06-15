import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu } from 'antd'
import {
  ScissorOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import Playground from './pages/Playground'
import RagPlayground from './pages/RagPlayground'

const { Header, Content } = Layout

const menuItems = [
  { key: '/', icon: <ScissorOutlined />, label: '切块演示' },
  { key: '/rag', icon: <SearchOutlined />, label: 'RAG 检索' },
]

function AppLayout() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{
        display: 'flex',
        alignItems: 'center',
        padding: '0 24px',
        gap: 24,
      }}>
        <span style={{
          color: '#fff',
          fontSize: 18,
          fontWeight: 600,
          whiteSpace: 'nowrap',
        }}>
          Wiki Playground
        </span>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ flex: 1, minWidth: 0 }}
        />
      </Header>
      <Content style={{ padding: 24 }}>
        <Routes>
          <Route path="/" element={<Playground />} />
          <Route path="/rag" element={<RagPlayground />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Content>
    </Layout>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AppLayout />
    </BrowserRouter>
  )
}

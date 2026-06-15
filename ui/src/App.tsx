import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from 'antd'
import Playground from './pages/Playground'

const { Header, Content } = Layout

export default function App() {
  return (
    <BrowserRouter>
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{
          color: '#fff',
          fontSize: 20,
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
        }}>
          📄 Wiki Chunk Playground
        </Header>
        <Content style={{ padding: 24 }}>
          <Routes>
            <Route path="/" element={<Playground />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Content>
      </Layout>
    </BrowserRouter>
  )
}

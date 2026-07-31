import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { Dashboard } from './pages/Dashboard';
import { LogsPage } from './pages/LogsPage';
import { RolesPage } from './pages/RolesPage';
import { DiffPage } from './pages/DiffPage';

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/logs" element={<LogsPage />} />
          <Route path="/roles" element={<RolesPage />} />
          <Route path="/diff" element={<DiffPage />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;

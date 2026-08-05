import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Home from './pages/Home';
import Services from './pages/Services';
import Health from './pages/Health';
import Metrics from './pages/Metrics';
import RequestHistory from './pages/RequestHistory';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/services" element={<Services />} />
        <Route path="/health" element={<Health />} />
        <Route path="/metrics" element={<Metrics />} />
        <Route path="/history" element={<RequestHistory />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}

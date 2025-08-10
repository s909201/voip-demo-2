import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import MainPage from './components/MainPage';
import CallHistory from './pages/CallHistory';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<MainPage />} />
        <Route path="/history" element={<CallHistory />} />
      </Routes>
    </Router>
  );
}

export default App;

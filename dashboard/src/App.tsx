import { useEffect, useState } from 'react';
import DashboardLayout from './components/DashboardLayout';
import { ensureAuthenticated, getUsername } from './utils/auth';

function App() {
  const [loading, setLoading] = useState(true);
  const [username, setUsername] = useState<string | null>(null);

  useEffect(() => {
    async function checkAuth() {
      // 检查是否已登录
      const loggedIn = await ensureAuthenticated();
      if (loggedIn) {
        setUsername(getUsername());
      }
      setLoading(false);
    }

    checkAuth();
  }, []);

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        fontSize: '18px',
      }}>
        加载中...
      </div>
    );
  }

  return (
    <>
      {username && (
        <div style={{
          position: 'fixed',
          top: '10px',
          right: '10px',
          background: '#fff',
          padding: '8px 16px',
          borderRadius: '4px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          zIndex: 1000,
          fontSize: '14px',
        }}>
          <span>用户: {username}</span>
          <button
            onClick={() => window.location.reload()}
            style={{
              marginLeft: '12px',
              padding: '4px 12px',
              background: '#ff4d4f',
              color: '#fff',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            重新登录
          </button>
        </div>
      )}
      <DashboardLayout />
    </>
  );
}

export default App;

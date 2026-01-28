/**
 * 简单的认证管理
 * 使用原生 prompt 弹窗进行登录
 */

const SESSION_KEY = 'agentflow_session_id';
const USERNAME_KEY = 'agentflow_username';

/**
 * 检查是否已登录
 */
export function isAuthenticated(): boolean {
  return !!localStorage.getItem(SESSION_KEY);
}

/**
 * 获取 Session ID
 */
export function getSessionId(): string | null {
  return localStorage.getItem(SESSION_KEY);
}

/**
 * 获取用户名
 */
export function getUsername(): string | null {
  return localStorage.getItem(USERNAME_KEY);
}

/**
 * 登录
 */
export async function login(): Promise<{ success: boolean; message: string }> {
  // 使用原生 prompt 弹窗获取用户名密码
  const username = prompt('请输入用户名:');
  if (!username) {
    return { success: false, message: '取消登录' };
  }

  const password = prompt('请输入密码:');
  if (!password) {
    return { success: false, message: '取消登录' };
  }

  try {
    const response = await fetch('/api/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();

    if (data.success && data.session_id) {
      // 存储 Session ID
      localStorage.setItem(SESSION_KEY, data.session_id);
      localStorage.setItem(USERNAME_KEY, username);
      return { success: true, message: data.message || '登录成功' };
    } else {
      return { success: false, message: data.message || '登录失败' };
    }
  } catch (error) {
    return { success: false, message: `登录失败: ${error}` };
  }
}

/**
 * 登出
 */
export function logout(): void {
  localStorage.removeItem(SESSION_KEY);
  localStorage.removeItem(USERNAME_KEY);
  window.location.reload();
}

/**
 * 确保已登录，如果未登录则显示登录弹窗
 */
export async function ensureAuthenticated(): Promise<boolean> {
  if (isAuthenticated()) {
    return true;
  }

  const result = await login();
  if (!result.success) {
    alert(`登录失败: ${result.message}`);
    return false;
  }

  return true;
}

/**
 * 为 fetch 请求添加认证头
 */
export function getAuthHeaders(): Record<string, string> {
  const sessionId = getSessionId();
  if (sessionId) {
    return {
      'Authorization': `Bearer ${sessionId}`,
    };
  }
  return {};
}

/**
 * 包装 fetch，自动添加认证头
 */
export async function authenticatedFetch(
  url: string,
  options?: RequestInit
): Promise<Response> {
  const headers = {
    ...options?.headers,
    ...getAuthHeaders(),
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // 如果返回 401，说明 Session 过期或无效
  if (response.status === 401) {
    logout();
  }

  return response;
}

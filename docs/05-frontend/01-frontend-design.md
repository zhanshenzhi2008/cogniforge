# CogniForge 前端设计文档

## 1. 技术栈选择

### 1.1 选型理由

| 技术 | 版本 | 选型理由 |
|-----|------|---------|
| **Vue** | 3.4+ | 组合式 API、响应式、直观易学 |
| **TypeScript** | 5+ | 强类型支持、IDE友好 |
| **Nuxt** | 3.14+ | SSR/SSG、内置 API 路由、自动导入 |
| **Tailwind CSS** | 3+ | 原子化CSS、快速开发、树摇优化 |
| **Pinia** | 2+ | 简单直观、TypeScript友好、模块化 |
| **Vue Query** | 5+ | 服务端状态管理、自动缓存 |
| **VueUse** | 10+ | 组合式工具库、浏览器API封装 |
| **Element Plus** | 2.6+ | 中文文档完善、组件丰富 |
| **Vue Flow** | 1.0+ | 工作流可视化编辑器首选 |
| **Zod** | 3+ | TypeScript原生支持、验证友好 |
| ** @vueuse/core** | 10+ | 组合式工具库 |

### 1.2 开发工具

```json
{
  "node": ">=20.0.0",
  "pnpm": ">=8.0.0",
  "typescript": ">=5.0.0",
  "eslint": ">=8.0.0",
  "prettier": ">=3.0.0"
}
```

---

## 2. 项目结构

### 2.1 目录结构

```
cogniforge-web/
├── public/                     # 静态资源
├── server/                     # Nuxt API 路由
│   └── api/                   # 服务端 API
│       ├── auth/              # 认证相关
│       ├── agents/            # Agent 相关
│       ├── workflows/         # 工作流相关
│       └── ...
│
├── src/
│   ├── assets/               # 资源文件
│   │   └── styles/           # 全局样式
│   │
│   ├── components/           # 全局组件 (自动导入)
│   │   ├── ui/              # 基础 UI 组件 (Element Plus)
│   │   ├── common/          # 通用业务组件
│   │   ├── forms/           # 表单组件
│   │   └── layouts/         # 布局组件
│   │
│   ├── composables/          # 组合式函数 (自动导入)
│   │   ├── useAuth.ts       # 认证
│   │   ├── useAgents.ts     # Agent 相关
│   │   ├── useChat.ts       # 聊天功能
│   │   └── ...
│   │
│   ├── features/            # 功能模块
│   │   ├── auth/           # 认证模块
│   │   ├── agents/         # Agent 模块
│   │   ├── workflows/      # 工作流模块
│   │   ├── knowledge/      # 知识库模块
│   │   ├── monitoring/     # 监控模块
│   │   ├── chat/          # 聊天模块
│   │   └── settings/      # 设置模块
│   │
│   ├── layouts/             # Nuxt 布局
│   │   ├── default.vue     # 默认布局
│   │   ├── auth.vue        # 认证布局 (无侧边栏)
│   │   └── dashboard.vue   # 控制台布局
│   │
│   ├── middleware/           # 中间件
│   │   ├── auth.ts         # 认证守卫
│   │   └── ...
│   │
│   ├── pages/               # 页面 (Nuxt 自动路由)
│   │   ├── index.vue       # 首页
│   │   ├── login.vue       # 登录页
│   │   ├── register.vue    # 注册页
│   │   └── dashboard/      # 控制台页面
│   │       ├── index.vue   # 概览
│   │       ├── agents/     # Agent 页面
│   │       │   ├── index.vue
│   │       │   ├── new.vue
│   │       │   └── [id].vue
│   │       ├── workflows/  # 工作流页面
│   │       ├── knowledge/  # 知识库页面
│   │       ├── models/     # 模型配置
│   │       ├── monitoring/ # 监控中心
│   │       └── settings/   # 设置页面
│   │
│   ├── plugins/             # Nuxt 插件
│   │   ├── api.ts          # API 客户端
│   │   ├── element.ts      # Element Plus
│   │   └── ...
│   │
│   ├── server/              # 服务端类型
│   │   └── api/            # API 类型定义
│   │
│   ├── stores/             # Pinia 状态管理
│   │   ├── auth.ts         # 认证状态
│   │   ├── agents.ts       # Agent 状态
│   │   ├── chat.ts         # 聊天状态
│   │   └── ...
│   │
│   ├── types/              # TypeScript 类型
│   ├── utils/              # 工具函数
│   ├── constants/          # 常量定义
│   ├── env.d.ts           # 环境变量类型
│   └── app.vue            # 根组件
│
├── nuxt.config.ts          # Nuxt 配置
├── tailwind.config.ts
├── tsconfig.json
├── package.json
└── .env.example            # 环境变量示例
```

### 2.2 路由结构

```
/                           # 首页 (重定向到 /dashboard 或 /login)
/login                      # 登录页
/register                   # 注册页
/dashboard                  # 控制台首页 (概览)
/dashboard/agents           # Agent 列表
/dashboard/agents/new      # 创建 Agent
/dashboard/agents/[id]     # Agent 详情/编辑
/dashboard/workflows       # 工作流列表
/dashboard/workflows/new   # 创建工作流
/dashboard/workflows/[id] # 工作流详情/编辑
/dashboard/knowledge      # 知识库列表
/dashboard/knowledge/[id]  # 知识库详情
/dashboard/models         # 模型配置
/dashboard/monitoring     # 监控中心
/dashboard/monitoring/logs # 请求日志
/dashboard/settings       # 设置
/dashboard/settings/team # 团队管理
/dashboard/settings/billing # 计费设置
```

---

## 3. 设计系统

### 3.1 色彩系统

```typescript
// src/lib/constants/colors.ts

export const colors = {
  // 主色 (CogniForge Blue)
  primary: {
    50: '#F0F9FF',
    100: '#E0F2FE',
    200: '#BAE6FD',
    300: '#7DD3FC',
    400: '#38BDF8',
    500: '#0EA5E9',    // 主色
    600: '#0284C7',
    700: '#0369A1',
    800: '#075985',
    900: '#0C4A6E',
  },
  
  // 辅助色
  success: {
    500: '#22C55E',
    600: '#16A34A',
  },
  warning: {
    500: '#F59E0B',
    600: '#D97706',
  },
  error: {
    500: '#EF4444',
    600: '#DC2626',
  },
  info: {
    500: '#8B5CF6',
    600: '#7C3AED',
  },
  
  // 中性色
  gray: {
    50: '#F9FAFB',
    100: '#F3F4F6',
    200: '#E5E7EB',
    300: '#D1D5DB',
    400: '#9CA3AF',
    500: '#6B7280',
    600: '#4B5563',
    700: '#374151',
    800: '#1F2937',
    900: '#111827',
  },
  
  // 暗色模式
  dark: {
    bg: '#0F172A',
    bgSecondary: '#1E293B',
    border: '#334155',
  }
}
```

### 3.2 字体系统

```typescript
// src/lib/constants/typography.ts

export const typography = {
  fontFamily: {
    sans: ['Inter', 'system-ui', 'sans-serif'],
    mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
  },
  
  fontSize: {
    xs: ['0.75rem', { lineHeight: '1rem' }],      // 12px
    sm: ['0.875rem', { lineHeight: '1.25rem' }],   // 14px
    base: ['1rem', { lineHeight: '1.5rem' }],      // 16px
    lg: ['1.125rem', { lineHeight: '1.75rem' }],  // 18px
    xl: ['1.25rem', { lineHeight: '1.75rem' }],    // 20px
    '2xl': ['1.5rem', { lineHeight: '2rem' }],     // 24px
    '3xl': ['1.875rem', { lineHeight: '2.25rem' }], // 30px
    '4xl': ['2.25rem', { lineHeight: '2.5rem' }],  // 36px
  },
  
  fontWeight: {
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
  }
}
```

### 3.3 间距系统

```typescript
// 基于 4px 网格
export const spacing = {
  0: '0',
  1: '0.25rem',   // 4px
  2: '0.5rem',    // 8px
  3: '0.75rem',   // 12px
  4: '1rem',      // 16px
  5: '1.25rem',   // 20px
  6: '1.5rem',    // 24px
  8: '2rem',      // 32px
  10: '2.5rem',   // 40px
  12: '3rem',     // 48px
  16: '4rem',     // 64px
}
```

---

## 3. 核心组件设计

### 3.1 布局组件

#### 3.1.1 主布局 (DashboardLayout)

```vue
<!-- src/layouts/dashboard.vue -->

<template>
  <div class="flex h-screen bg-gray-50">
    <!-- 侧边栏 -->
    <aside class="w-64 bg-white border-r border-gray-200">
      <Sidebar />
    </aside>

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 顶部导航 -->
      <header class="h-16 bg-white border-b border-gray-200 flex items-center px-6">
        <TopNav />
      </header>

      <!-- 内容 -->
      <main class="flex-1 overflow-auto p-6">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
// 布局逻辑
</script>
```

#### 3.1.2 侧边栏导航

```
┌─────────────────────────────────────────────┐
│  [Logo] CogniForge                    [🔔][?]│
├─────────────────────────────────────────────┤
│                                             │
│  📊 概览                                    │
│                                             │
│  ─────────────────────────────────────────  │
│                                             │
│  🤖 Agent                      ▶            │
│     ├─ 我的Agent                           │
│     ├─ 工具市场                            │
│     └─ 配置                                │
│                                             │
│  🔄 工作流                    ▶             │
│     ├─ 我的工作流                          │
│     └─ 模板市场                            │
│                                             │
│  📚 知识库                    ▶             │
│     ├─ 我的知识库                          │
│     └─ 文档                               │
│                                             │
│  💬 聊天                              │
│                                             │
│  🤖 模型                              │
│                                             │
│  📈 监控                              │
│     ├─ 概览                               │
│     ├─ 日志                               │
│     └─ 告警                               │
│                                             │
│  ─────────────────────────────────────────  │
│                                             │
│  ⚙️ 设置                      ▶             │
│     ├─ 账户                                │
│     ├─ 团队                                │
│     ├─ API密钥                            │
│     └─ 计费                                │
│                                             │
└─────────────────────────────────────────────┘
```

### 3.2 Playground/Chat 组件

```vue
<!-- src/features/chat/ChatWindow.vue -->

<template>
  <div class="flex h-full">
    <!-- 左侧：配置面板 -->
    <div class="w-80 bg-white border-r border-gray-200 p-4">
      <ModelSelector v-model="selectedModel" />
      <ParamsPanel v-model="params" />
    </div>

    <!-- 右侧：对话区 -->
    <div class="flex-1 flex flex-col">
      <!-- 消息列表 -->
      <div class="flex-1 overflow-auto p-4">
        <MessageList :messages="chatStore.messages" />
      </div>

      <!-- 输入区 -->
      <div class="border-t border-gray-200 p-4">
        <MessageInput @send="handleSend" :loading="chatStore.streaming" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useChatStore } from '@/stores/chat'
import type { Model, ChatParams } from '@/types'

const chatStore = useChatStore()

const selectedModel = ref<Model | null>(null)
const params = ref<ChatParams>({
  temperature: 0.7,
  maxTokens: 1000,
  topP: 1.0,
})

const handleSend = async (content: string) => {
  if (!selectedModel.value) return
  await chatStore.sendMessage(content, selectedModel.value.id)
}
</script>
```

### 3.3 工作流编辑器

```vue
<!-- src/features/workflows/WorkflowEditor.vue -->

<template>
  <div class="h-full flex">
    <!-- 左侧：节点面板 -->
    <div class="w-56 bg-white border-r border-gray-200 p-3">
      <NodePalette @drag-start="handleDragStart" />
    </div>

    <!-- 中间：画布 -->
    <div class="flex-1">
      <VueFlow
        v-model="nodes"
        v-model:edges="edges"
        :node-types="nodeTypes"
        @connect="handleConnect"
      >
        <Controls />
        <Background />
        <MiniMap />
      </VueFlow>
    </div>

    <!-- 右侧：属性面板 -->
    <div class="w-80 bg-white border-l border-gray-200 p-3">
      <PropertiesPanel
        :node="selectedNode"
        @update="updateNode"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import '@vue-flow/core/dist/style.css'

const { nodes, edges, onConnect } = useVueFlow()

const nodeTypes = {
  start: StartNode,
  end: EndNode,
  llm: LLMNode,
  agent: AgentNode,
  condition: ConditionNode,
  http: HttpNode,
  search: SearchNode,
}

const selectedNode = ref(null)

const handleConnect = (connection: Connection) => {
  edges.value.push({
    id: crypto.randomUUID(),
    source: connection.source,
    target: connection.target,
  })
}
</script>
```

### 3.4 监控仪表板

```vue
<!-- src/features/monitoring/MonitoringDashboard.vue -->

<template>
  <div class="space-y-6">
    <!-- 概览卡片 -->
    <div class="grid grid-cols-4 gap-4">
      <StatCard
        title="总请求"
        value="1.2M"
        change="+12%"
        trend="up"
      />
      <StatCard
        title="成功率"
        value="99.8%"
        change="-0.1%"
        trend="down"
      />
      <StatCard
        title="平均延迟"
        value="850ms"
        change="-5%"
        trend="up"
      />
      <StatCard
        title="本月成本"
        value="¥8,567"
        change="+8%"
        trend="down"
      />
    </div>

    <!-- 图表区域 -->
    <div class="grid grid-cols-2 gap-6">
      <el-card title="请求量趋势">
        <RequestTrendChart />
      </el-card>
      <el-card title="延迟分布">
        <LatencyChart />
      </el-card>
    </div>

    <!-- 表格区域 -->
    <el-card title="Top模型调用">
      <ModelUsageTable />
    </el-card>
  </div>
</template>

<script setup lang="ts">
// 监控逻辑
</script>
```

---

## 4. 状态管理

### 4.1 Pinia Store

```typescript
// src/stores/auth.ts

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { useApi } from '@/composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  // Actions
  async function login(email: string, password: string) {
    loading.value = true
    try {
      const api = useApi()
      const response = await api.post('/auth/login', { email, password })
      token.value = response.token
      user.value = response.user
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    const api = useApi()
    await api.post('/auth/logout')
    token.value = null
    user.value = null
  }

  async function fetchUser() {
    if (!token.value) return
    loading.value = true
    try {
      const api = useApi()
      const response = await api.get<User>('/auth/me')
      user.value = response
    } catch {
      token.value = null
      user.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    token,
    loading,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    fetchUser,
  }
})
```

### 4.2 聊天状态 Store

```typescript
// src/stores/chat.ts

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Message, ChatSession } from '@/types'

export const useChatStore = defineStore('chat', () => {
  // State
  const sessions = ref<ChatSession[]>([])
  const currentSession = ref<ChatSession | null>(null)
  const messages = ref<Message[]>([])
  const loading = ref(false)
  const streaming = ref(false)

  // Getters
  const currentMessages = computed(() => messages.value)

  // Actions
  function addMessage(message: Message) {
    messages.value.push(message)
  }

  function updateLastMessage(content: string) {
    if (messages.value.length > 0) {
      const lastMessage = messages.value[messages.value.length - 1]
      lastMessage.content = content
    }
  }

  function clearMessages() {
    messages.value = []
  }

  async function sendMessage(content: string, agentId: string) {
    // 添加用户消息
    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: new Date(),
    }
    addMessage(userMessage)

    // 添加空的助手消息占位
    const assistantMessage: Message = {
      id: crypto.randomUUID(),
      role: 'assistant',
      content: '',
      timestamp: new Date(),
    }
    addMessage(assistantMessage)

    streaming.value = true
    // 调用流式 API
    // ...
    streaming.value = false
  }

  return {
    sessions,
    currentSession,
    messages,
    loading,
    streaming,
    currentMessages,
    addMessage,
    updateLastMessage,
    clearMessages,
    sendMessage,
  }
})
```

### 4.3 Vue Query 客户端

```typescript
// src/plugins/vue-query.ts

import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query'

export default defineNuxtPlugin((nuxtApp) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000,    // 5 分钟
        gcTime: 30 * 60 * 1000,      // 30 分钟
        retry: 3,
        refetchOnWindowFocus: false,
      },
    },
  })

  nuxtApp.vueApp.use(VueQueryPlugin, { queryClient })
})
```

```typescript
// src/composables/useAgents.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import type { Agent, CreateAgentInput } from '@/types'
import { useApi } from './useApi'

export function useAgents() {
  const api = useApi()

  return useQuery({
    queryKey: ['agents'],
    queryFn: async () => {
      const response = await api.get<Agent[]>('/v1/agents')
      return response
    },
  })
}

export function useAgent(id: string) {
  const api = useApi()

  return useQuery({
    queryKey: ['agent', id],
    queryFn: async () => {
      const response = await api.get<Agent>(`/v1/agents/${id}`)
      return response
    },
    enabled: !!id,
  })
}

export function useCreateAgent() {
  const api = useApi()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateAgentInput) => {
      const response = await api.post<Agent>('/v1/agents', data)
      return response
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}
```

---

## 5. API 客户端

### 5.1 API 实例

```typescript
// src/composables/useApi.ts

import { useAuthStore } from '@/stores/auth'

export function useApi() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const baseURL = config.public.apiBase as string || '/api'

  async function request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    }

    if (authStore.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${authStore.token}`
    }

    const response = await fetch(`${baseURL}${endpoint}`, {
      ...options,
      headers,
    })

    if (response.status === 401) {
      await authStore.logout()
      navigateTo('/login')
      throw new Error('Unauthorized')
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({}))
      throw new Error(error.message || 'Request failed')
    }

    return response.json()
  }

  function get<T>(endpoint: string): Promise<T> {
    return request<T>(endpoint, { method: 'GET' })
  }

  function post<T>(endpoint: string, data?: unknown): Promise<T> {
    return request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  function put<T>(endpoint: string, data?: unknown): Promise<T> {
    return request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  function del<T>(endpoint: string): Promise<T> {
    return request<T>(endpoint, { method: 'DELETE' })
  }

  // 流式请求
  async function stream(
    endpoint: string,
    data: unknown,
    onChunk: (content: string) => void
  ): Promise<void> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    if (authStore.token) {
      headers['Authorization'] = `Bearer ${authStore.token}`
    }

    const response = await fetch(`${baseURL}${endpoint}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(data),
    })

    if (!response.ok) {
      throw new Error('Stream request failed')
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (!reader) return

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      const chunk = decoder.decode(value)
      onChunk(chunk)
    }
  }

  return {
    get,
    post,
    put,
    del,
    stream,
  }
}
```

### 5.2 API 模块

```typescript
// src/composables/useAgents.ts

import type { Agent, CreateAgentInput, UpdateAgentInput } from '@/types'
import { useApi } from './useApi'

export function useAgentsApi() {
  const api = useApi()

  return {
    list: () => api.get<Agent[]>('/v1/agents'),

    get: (id: string) => api.get<Agent>(`/v1/agents/${id}`),

    create: (data: CreateAgentInput) =>
      api.post<Agent>('/v1/agents', data),

    update: (id: string, data: UpdateAgentInput) =>
      api.put<Agent>(`/v1/agents/${id}`, data),

    delete: (id: string) =>
      api.del<void>(`/v1/agents/${id}`),

    chat: (id: string, messages: Message[]) =>
      api.post<ChatResponse>(`/v1/agents/${id}/chat`, { messages }),

    streamChat: (id: string, messages: Message[], onChunk: (content: string) => void) =>
      api.stream(`/v1/agents/${id}/chat/stream`, { messages }, onChunk),
  }
}
```

---

## 6. 表单处理

### 6.1 创建 Agent 表单

```vue
<!-- src/features/agents/components/CreateAgentForm.vue -->

<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="100px"
  >
    <el-form-item label="名称" prop="name">
      <el-input v-model="form.name" placeholder="请输入 Agent 名称" />
    </el-form-item>

    <el-form-item label="描述" prop="description">
      <el-input
        v-model="form.description"
        type="textarea"
        :rows="3"
        placeholder="请输入描述"
      />
    </el-form-item>

    <el-form-item label="模型" prop="model">
      <el-select v-model="form.model" placeholder="请选择模型">
        <el-option label="GPT-4o" value="gpt-4o" />
        <el-option label="GPT-4o-mini" value="gpt-4o-mini" />
        <el-option label="Claude 3.5" value="claude-3.5-sonnet" />
      </el-select>
    </el-form-item>

    <el-form-item label="系统提示" prop="systemPrompt">
      <el-input
        v-model="form.systemPrompt"
        type="textarea"
        :rows="6"
        placeholder="请输入系统提示词"
      />
    </el-form-item>

    <el-form-item label="工具" prop="tools">
      <el-checkbox-group v-model="form.tools">
        <el-checkbox label="web_search">网页搜索</el-checkbox>
        <el-checkbox label="code_executor">代码执行</el-checkbox>
        <el-checkbox label="file_reader">文件读取</el-checkbox>
      </el-checkbox-group>
    </el-form-item>

    <el-form-item>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        创建
      </el-button>
      <el-button @click="handleReset">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useCreateAgent } from '@/composables/useAgents'

const formRef = ref<FormInstance>()
const createAgent = useCreateAgent()
const loading = ref(false)

const form = reactive({
  name: '',
  description: '',
  model: 'gpt-4o',
  systemPrompt: '',
  tools: [] as string[],
})

const rules: FormRules = {
  name: [
    { required: true, message: '名称不能为空', trigger: 'blur' },
    { max: 100, message: '名称不能超过100个字符', trigger: 'blur' },
  ],
  model: [
    { required: true, message: '请选择模型', trigger: 'change' },
  ],
  systemPrompt: [
    { max: 10000, message: '系统提示不能超过10000个字符', trigger: 'blur' },
  ],
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await createAgent.mutateAsync(form)
        ElMessage.success('创建成功')
      } catch (error) {
        ElMessage.error('创建失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const handleReset = () => {
  formRef.value?.resetFields()
}
</script>
```

### 6.2 Zod 验证规则

```typescript
// src/utils/validators/agent.ts

import { z } from 'zod'

export const agentSchema = z.object({
  name: z.string()
    .min(1, '名称不能为空')
    .max(100, '名称不能超过100个字符'),

  description: z.string()
    .max(500, '描述不能超过500个字符')
    .optional(),

  model: z.string()
    .min(1, '请选择模型'),

  systemPrompt: z.string()
    .max(10000, '系统提示不能超过10000个字符'),

  tools: z.array(z.string()),

  memory: z.object({
    type: z.enum(['short_term', 'long_term']),
    maxTurns: z.number().min(1).max(100),
  }),
})

export type AgentFormData = z.infer<typeof agentSchema>
```

---

## 8. 响应式设计

### 8.1 断点定义

```typescript
// tailwind.config.ts

export default {
  theme: {
    extend: {
      screens: {
        'xs': '475px',    // 超小屏幕
        'sm': '640px',    // 小屏幕
        'md': '768px',    // 中屏幕
        'lg': '1024px',   // 大屏幕
        'xl': '1280px',   // 超大屏幕
        '2xl': '1536px',  // 特大屏幕
      },
    },
  },
}
```

### 8.2 响应式组件示例

```tsx
// 响应式侧边栏
<aside className="
  hidden lg:block w-64        // 大屏幕显示
  fixed inset-y-0 left-0      // 固定定位
  z-50                        // 层级
">
  <Sidebar />
</aside>

// 移动端菜单按钮
<Button 
  className="lg:hidden"
  variant="ghost"
  size="icon"
  onClick={() => setMobileMenuOpen(true)}
>
  <MenuIcon />
</Button>
```

---

## 8. 性能优化

### 8.1 优化策略

| 优化点 | 策略 |
|-------|------|
| **代码分割** | Nuxt 自动路由分割 |
| **图片优化** | @nuxt/image 自动优化 |
| **API缓存** | Vue Query 自动缓存 |
| **虚拟列表** | 使用 vue-virtual-scroller 处理大数据 |
| **骨架屏** | loading 状态展示骨架 |
| **懒加载** | defineAsyncComponent 组件 |

### 8.2 懒加载示例

```vue
<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

// 工作流编辑器懒加载
const WorkflowEditor = defineAsyncComponent(() =>
  import('@/features/workflows/WorkflowEditor.vue')
)

// 带 loading
const WorkflowEditorWithLoading = defineAsyncComponent({
  loader: () => import('@/features/workflows/WorkflowEditor.vue'),
  loadingComponent: () => import('@/components/EditorSkeleton.vue'),
  delay: 200,
})

// 代码编辑器懒加载
const CodeEditor = defineAsyncComponent(() =>
  import('@/components/common/CodeEditor.vue')
)
</script>
```

---

## 9. 测试策略

### 9.1 测试工具

| 工具 | 用途 |
|-----|------|
| **Vitest** | 单元测试 |
| **Vue Test Utils** | 组件测试 |
| **Playwright** | E2E测试 |
| **MSW** | API Mock |

### 9.2 测试示例

```typescript
// __tests__/components/Button.spec.ts

import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import { Button } from '@/components/ui/Button.vue'

describe('Button', () => {
  it('renders correctly', () => {
    const wrapper = mount(Button, {
      slots: { default: 'Click me' }
    })
    expect(wrapper.text()).toContain('Click me')
  })

  it('calls onClick handler', () => {
    const handleClick = vi.fn()
    const wrapper = mount(Button, {
      props: { onClick: handleClick },
      slots: { default: 'Click me' }
    })
    wrapper.trigger('click')
    expect(handleClick).toHaveBeenCalledTimes(1)
  })
})
```

```typescript
// __tests__/composables/useAuth.spec.ts

import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { vi } from 'vitest'

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('login successfully', async () => {
    const store = useAuthStore()
    const mockApi = vi.fn().mockResolvedValue({
      token: 'mock-token',
      user: { id: '1', name: 'Test' },
    })

    // Mock API
    vi.stubGlobal('useApi', () => ({
      post: mockApi,
    }))

    await store.login('test@example.com', 'password')

    expect(store.token).toBe('mock-token')
    expect(store.user?.name).toBe('Test')
  })
})
```

---

**文档版本**: v1.0  
**最后更新**: 2026-03-16  
**维护团队**: orjrs

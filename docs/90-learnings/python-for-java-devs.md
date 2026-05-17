# Python 入门指南（Java 开发者视角）

本文档为 Java 开发者快速上手 Cogniforge 项目中的 Python 服务。

## 目录

- [基础语法对照](#基础语法对照)
- [数据类型对比](#数据类型对比)
- [函数与方法](#函数与方法)
- [类与面向对象](#类与面向对象)
- [异常处理](#异常处理)
- [类型提示](#类型提示)
- [项目结构](#项目结构)
- [核心代码解读](#核心代码解读)

---

## 基础语法对照

### 注释

```python
# 单行注释

"""
多行注释（字符串常用于文档）
"""

'''
也可以用单引号
'''
```

**Java 对比：**
```java
// 单行注释

/*
 * 多行注释
 */
```

---

### 变量声明

```python
# Python 动态类型，不需要声明类型
name = "cogniforge"
age = 25
is_active = True

# 可以添加类型提示（推荐）
name: str = "cogniforge"
age: int = 25
is_active: bool = True
```

**Java 对比：**
```java
String name = "cogniforge";
int age = 25;
boolean isActive = true;
```

**关键差异：**
- Python 用 `_` 命名（下划线），Java 用 camelCase
- Python 用 `True/False`，Java 用 `true/false`

---

### 字符串

```python
# 基本用法
name = "Cogniforge"
greeting = f"Hello, {name}!"  # f-string 格式化

# 多行字符串
text = """
    这是多行文本
    可以跨多行
"""

# 字符串方法
name.upper()      # "COGNIFORGE"
name.lower()      # "cogniforge"
name.strip()      # 去除首尾空格
"hello" + "world" # 字符串拼接
```

**Java 对比：**
```java
String name = "Cogniforge";
String greeting = "Hello, " + name + "!";  // 或 String.format()

String text = """
    这是多行文本
""";
```

---

### 条件判断

```python
age = 18

if age >= 18:
    print("成年人")
elif age >= 6:
    print("未成年人")
else:
    print("儿童")

# Python 用 elif，Java 用 else if
```

**Java 对比：**
```java
if (age >= 18) {
    System.out.println("成年人");
} else if (age >= 6) {
    System.out.println("未成年人");
} else {
    System.out.println("儿童");
}
```

---

### 循环

```python
# for 循环（类似 Java 的增强 for）
fruits = ["apple", "banana", "orange"]

for fruit in fruits:
    print(fruit)

# 带索引
for i, fruit in enumerate(fruits):
    print(f"{i}: {fruit}")

# range 范围循环（类似 Java 的普通 for）
for i in range(5):      # 0, 1, 2, 3, 4
    print(i)

for i in range(1, 6):   # 1, 2, 3, 4, 5
    print(i)

# while 循环
count = 0
while count < 5:
    count += 1
```

**Java 对比：**
```java
String[] fruits = {"apple", "banana", "orange"};

for (String fruit : fruits) {
    System.out.println(fruit);
}

for (int i = 0; i < 5; i++) {
    System.out.println(i);
}
```

---

### 缩进

**Python 用缩进表示代码块，Java 用 `{}`。这是最重要的区别！**

```python
if True:
    print("这行缩进了")
    print("这行也缩进了")  # 同一代码块
print("这行没有缩进")      # 不属于 if
```

---

## 数据类型对比

### 集合类型

| Python | Java | 说明 |
|--------|------|------|
| `list` | `List` | 有序、可重复 |
| `dict` | `Map` | 键值对 |
| `set` | `Set` | 无序、不重复 |
| `tuple` | - | 不可变列表 |

### List（列表）

```python
# 创建
fruits = ["apple", "banana", "orange"]
numbers = list([1, 2, 3])

# 常用操作
fruits.append("grape")      # 添加
fruits[0]                   # 获取（0-indexed）
fruits[1:3]                  # 切片 [start:end]，类似 subList
len(fruits)                 # 长度，类似 size()
fruits.remove("apple")      # 移除
```

**Java 对比：**
```java
List<String> fruits = new ArrayList<>();
fruits.add("apple");
fruits.get(0);
fruits.subList(1, 3);
fruits.size();
fruits.remove("apple");
```

### Dict（字典）

```python
# 创建
user = {"name": "Alice", "age": 25}

# 常用操作
user["name"]                # 获取
user["email"] = "a@b.com"   # 设置
user.get("name")            # 安全获取，不存在返回 None
user.keys()                 # 所有键
user.values()               # 所有值
user.items()                # 键值对
```

**Java 对比：**
```java
Map<String, Object> user = new HashMap<>();
user.get("name");
user.put("email", "a@b.com");
user.getOrDefault("name", "default");
user.keySet();
user.values();
user.entrySet();
```

### Optional（空值处理）

```python
# Python 用 None 表示空值
name = None
name = user.get("name")     # 可能返回 None

# 空值检查
if name is not None:
    print(name)

# 链式调用（Python 3.10+）
result = user.get("profile", {}).get("avatar")

# 或用默认值
name = user.get("name", "Anonymous")
```

**Java 对比：**
```java
String name = null;
Optional<String> nameOpt = Optional.ofNullable(user.get("name"));

if (nameOpt.isPresent()) {
    System.out.println(nameOpt.get());
}

String name = user.getOrDefault("name", "Anonymous");
```

---

## 函数与方法

### 函数定义

```python
# 基本函数
def greet(name: str) -> str:
    """这是函数的文档字符串"""
    return f"Hello, {name}!"

# 默认参数
def greet(name: str, greeting: str = "Hello") -> str:
    return f"{greeting}, {name}!"

# 可变参数
def sum(*numbers):
    total = 0
    for n in numbers:
        total += n
    return total

sum(1, 2, 3)  # 6

# 关键字参数
def create_user(name: str, age: int, city: str = "Beijing"):
    pass

create_user(name="Alice", age=25)  # 关键字参数，顺序无关
```

**Java 对比：**
```java
public String greet(String name) {
    return "Hello, " + name + "!";
}

public String greet(String name, String greeting) {
    return greeting + ", " + name + "!";
}

public int sum(int... numbers) {
    int total = 0;
    for (int n : numbers) {
        total += n;
    }
    return total;
}
```

### Lambda 表达式

```python
# 单行匿名函数
square = lambda x: x * x

# 带参数
add = lambda a, b: a + b

# 在高阶函数中使用
numbers = [1, 2, 3, 4, 5]
squared = list(map(lambda x: x * x, numbers))
evens = list(filter(lambda x: x % 2 == 0, numbers))
```

**Java 对比：**
```java
Function<Integer, Integer> square = x -> x * x;

BiFunction<Integer, Integer, Integer> add = (a, b) -> a + b;

List<Integer> squared = numbers.stream()
    .map(x -> x * x)
    .collect(Collectors.toList());
```

---

## 类与面向对象

### 类定义

```python
class DocumentProcessor:
    """文档处理服务类"""

    def __init__(self, chunk_size: int = 512, overlap: int = 50):
        """
        构造方法（Python 用 __init__）
        self 相当于 Java 的 this
        """
        self.chunk_size = chunk_size      # 实例变量
        self.overlap = overlap
        self._internal_state = True      # 下划线前缀表示私有

    def process(self, file_path: str) -> bool:
        """实例方法，第一个参数是 self"""
        return True

    @staticmethod
    def validate_path(path: str) -> bool:
        """静态方法，不需要 self"""
        return path is not None

    @classmethod
    def create_default(cls):
        """类方法，第一个参数是类本身"""
        return cls(chunk_size=1024)
```

**Java 对比：**
```java
public class DocumentProcessor {

    private int chunkSize;
    private int overlap;
    private boolean internalState;

    public DocumentProcessor(int chunkSize, int overlap) {
        this.chunkSize = chunkSize;
        this.overlap = overlap;
        this.internalState = true;
    }

    public boolean process(String filePath) {
        return true;
    }

    public static boolean validatePath(String path) {
        return path != null;
    }

    public static DocumentProcessor createDefault() {
        return new DocumentProcessor(1024, 50);
    }
}
```

### 继承

```python
class BaseParser:
    """基类"""
    def parse(self, file_path: str):
        raise NotImplementedError  # 抽象方法


class PDFParser(BaseParser):
    """子类继承"""
    SUPPORTED_EXTENSIONS = ['.pdf']  # 类变量

    def __init__(self):
        super().__init__()          # 调用父类构造

    def parse(self, file_path: str):  # 实现抽象方法
        # 实现代码
        pass
```

**Java 对比：**
```java
public abstract class BaseParser {
    public abstract void parse(String filePath);
}

public class PDFParser extends BaseParser {
    public static final String[] SUPPORTED_EXTENSIONS = {".pdf"};

    public PDFParser() {
        super();
    }

    @Override
    public void parse(String filePath) {
        // 实现代码
    }
}
```

### 数据类（Dataclass）

```python
from dataclasses import dataclass
from typing import Optional

@dataclass
class ChunkResult:
    """简单的数据容器，类似 Java 的 Lombok @Data"""
    chunk_id: str
    content: str
    index: int
    metadata: dict = None  # 默认值

# 自动生成 __init__, __repr__, __eq__ 等方法
result = ChunkResult("id1", "content", 0)
print(result.chunk_id)  # 直接访问属性
```

**Java 对比：**
```java
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ChunkResult {
    private String chunkId;
    private String content;
    private int index;
    private Map<String, Object> metadata;
}
```

---

## 异常处理

### 基本语法

```python
try:
    result = processor.process_document(file_path)
except ValueError as e:
    print(f"参数错误: {e}")
except Exception as e:
    print(f"未知错误: {e}")
    raise  # 重新抛出
else:
    # 没有异常时执行
    print("处理成功")
finally:
    # 无论是否异常都执行
    print("清理资源")
```

**Java 对比：**
```java
try {
    result = processor.processDocument(filePath);
} catch (ValueErrorException e) {
    System.out.println("参数错误: " + e.getMessage());
} catch (Exception e) {
    System.out.println("未知错误: " + e.getMessage());
    throw e;
} finally {
    System.out.println("清理资源");
}
```

### 自定义异常

```python
class ProcessingError(Exception):
    """自定义异常"""
    def __init__(self, message: str, code: int = 500):
        self.message = message
        self.code = code
        super().__init__(self.message)

# 抛出异常
raise ProcessingError("处理失败", 500)
```

**Java 对比：**
```java
public class ProcessingError extends RuntimeException {
    private final int code;

    public ProcessingError(String message, int code) {
        super(message);
        this.code = code;
    }
}

throw new ProcessingError("处理失败", 500);
```

---

## 类型提示

Python 3.5+ 支持类型提示（类似 TypeScript），但运行时不做检查。

### 基本用法

```python
from typing import List, Dict, Optional, Union, Callable

# 函数参数和返回值
def process(
    texts: List[str],
    config: Dict[str, int],
    callback: Optional[Callable] = None
) -> List[float]:
    ...

# Union 类型（类似泛型）
user_id: Union[str, int] = "user123"

# 类型别名
Vector = List[float]
Matrix = List[List[float]]

# 泛型
from typing import TypeVar
T = TypeVar('T')

def first(items: List[T]) -> Optional[T]:
    return items[0] if items else None
```

### Pydantic 模型（推荐）

```python
from pydantic import BaseModel

class SearchRequest(BaseModel):
    """请求模型，自动验证和序列化"""
    query: str
    collection_name: str
    top_k: int = 5           # 带默认值
    min_score: float = 0.0

    class Config:
        # Pydantic 配置
        json_schema_extra = {...}
```

**Java 对比：**
```java
public class SearchRequest {
    private String query;
    private String collectionName;
    private int topK = 5;
    private float minScore = 0.0F;

    // getters, setters, 或用 Lombok @Data
}
```

---

## 项目结构

```
llm/knowledge/
├── app/
│   ├── main.py          # FastAPI 入口，定义路由
│   └── routes.py        # API 路由（可选分离）
├── services/
│   └── document_processor.py  # 核心业务逻辑
├── parsers/             # 文档解析器
│   ├── base.py
│   ├── pdf_parser.py
│   ├── docx_parser.py
│   └── txt_parser.py
├── splitters/           # 文本分块
│   ├── base.py
│   └── recursive_splitter.py
├── embedding/           # 向量化
│   ├── base.py
│   ├── openai_embedder.py
│   └── local_embedder.py
├── vector_store/        # 向量数据库
│   ├── base.py
│   └── pgvector_store.py
├── utils/
│   └── __init__.py
├── requirements.txt     # 依赖
└── main.py              # 启动入口（可选）
```

---

## 核心代码解读

### FastAPI 路由（类似 Spring MVC）

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Knowledge Service", version="1.0.0")

class ProcessRequest(BaseModel):
    file_path: str
    document_id: str
    collection_name: str

@app.post("/api/knowledge/process")
async def process_document(request: ProcessRequest):
    """POST 请求处理"""
    if not processor:
        raise HTTPException(status_code=503, detail="Service unavailable")

    result = processor.process_document(...)
    return {"success": result.success, "chunks": len(result.chunks)}
```

**Java 对比（Spring Boot）：**
```java
@RestController
@RequestMapping("/api/knowledge")
public class KnowledgeController {

    @PostMapping("/process")
    public ResponseEntity<?> processDocument(@RequestBody ProcessRequest request) {
        if (processor == null) {
            return ResponseEntity.status(503).body("Service unavailable");
        }
        ProcessingResult result = processor.processDocument(...);
        return ResponseEntity.ok(Map.of("success", result.isSuccess(), "chunks", result.getChunks().size()));
    }
}
```

### 依赖注入（类似 Spring）

Python 没有内置 DI，通常用函数参数或全局变量：

```python
# 方式1：全局变量（简单，不推荐）
processor: Optional[DocumentProcessor] = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global processor
    processor = DocumentProcessor()
    yield
    # 清理

# 方式2：依赖注入（推荐，类似 Spring）
from fastapi import Depends

def get_processor():
    return processor

@app.get("/search")
async def search(query: str, proc = Depends(get_processor)):
    return proc.search(query)
```

**Java 对比（Spring）：**
```java
@Service
public class KnowledgeService {
    private final DocumentProcessor processor;

    @Autowired
    public KnowledgeService(DocumentProcessor processor) {
        this.processor = processor;
    }
}
```

### 上下文管理器（类似 try-with-resources）

```python
# Python 的 with 语句
with tempfile.NamedTemporaryFile(delete=False) as tmp:
    shutil.copyfileobj(file.file, tmp)
    temp_path = tmp.name
# 文件自动关闭，不需要 finally

# 自定义上下文管理器
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时
    processor = DocumentProcessor()
    yield
    # 关闭时
    processor.disconnect()
```

**Java 对比：**
```java
// try-with-resources
try (FileInputStream fis = new FileInputStream(file)) {
    // 使用 fis
} // 自动关闭

// Spring 的 @PostConstruct 和 @PreDestroy
@PreDestroy
public void cleanup() {
    processor.disconnect();
}
```

### 异步编程（类似 Java CompletableFuture）

```python
import asyncio

# 同步函数
def sync_function():
    return "result"

# 异步函数
async def async_function():
    await asyncio.sleep(1)  # 模拟 IO 操作
    return "result"

# 调用异步函数
async def main():
    result = await async_function()
    print(result)

asyncio.run(main())
```

**FastAPI 中的异步：**
```python
# FastAPI 自动处理异步
@app.get("/items/{item_id}")
async def read_item(item_id: int):
    # 同步操作（阻塞）
    item = db.query(item_id)
    return item

@app.get("/items/{item_id}")
async def read_item(item_id: int):
    # 异步操作（非阻塞）
    item = await async_db_query(item_id)
    return item
```

**Java 对比：**
```java
// 同步
public Item readItem(int itemId) {
    return db.query(itemId);
}

// 异步
public CompletableFuture<Item> readItemAsync(int itemId) {
    return CompletableFuture.supplyAsync(() -> db.query(itemId));
}
```

---

## 常用命令

```bash
# 安装依赖
pip install -r requirements.txt

# 运行服务
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# 运行测试（如果有）
pytest

# 代码格式化
black .

# 类型检查
mypy .
```

---

## 快速参考

| Java | Python |
|------|--------|
| `void` | `None` |
| `null` | `None` |
| `true/false` | `True/False` |
| `String` | `str` |
| `int`, `long` | `int` |
| `double`, `float` | `float` |
| `List<T>` | `List[T]` |
| `Map<K,V>` | `Dict[K, V]` |
| `{}` 包围代码块 | 缩进表示代码块 |
| `camelCase` | `snake_case` |
| `public/private` | `_` 前缀表示私有 |
| `@Override` | 不需要（自动检查） |
| `interface` | `ABC` + `@abstractmethod` |
| `throws` | `raise` |

---

## 推荐资源

- [Python 官方教程](https://docs.python.org/zh-cn/3/tutorial/)
- [FastAPI 文档](https://fastapi.tiangolo.com/zh/)
- [Pydantic 文档](https://docs.pydantic.dev/)
- [Python 3.11 新特性](https://docs.python.org/zh-cn/3/whatsnew/3.11.html)

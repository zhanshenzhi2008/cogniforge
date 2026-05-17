# Java 开发者转向 Go 的架构实践指南

> 本文专为有 Java/Spring Boot 背景的开发者编写，对比两者的设计理念与实践。

---

## 目录

1. [类型系统：继承 vs 组合](#1-类型系统继承-vs-组合)
2. [分层架构：Entity/DTO/VO 的取舍](#2-分层架构entitydtovo-的取舍)
3. [包管理：Maven/Gradle vs Go Modules](#3-包管理mavengradle-vs-go-modules)
4. [依赖注入：Spring IOC vs 构造函数注入](#4-依赖注入spring-ioc-vs-构造函数注入)
5. [错误处理：Exception vs Error](#5-错误处理exception-vs-error)
6. [并发模型：Thread vs Goroutine](#6-并发模型thread-vs-goroutine)
7. [泛型：Java Generics vs Go Generics](#7-泛型java-generics-vs-go-generics)
8. [ORM：JPA/Hibernate vs GORM](#8-ormjpahibernate-vs-gorm)
9. [接口设计：Interface 的哲学](#9-接口设计interface-的哲学)
10. [实战建议](#10-实战建议)
11. [测试实践：JUnit vs testing](#11-测试实践junit-vs-testing)
12. [缓存设计模式：Redis 实践](#12-缓存设计模式redis-实践)
13. [微服务架构：Spring Cloud vs Go 微服务](#13-微服务架构spring-cloud-vs-go-微服务)
14. [Docker 与 Kubernetes 部署](#14-docker-与-kubernetes-部署)
15. [Go 设计模式](#15-go-设计模式)
16. [性能优化技巧](#16-性能优化技巧)
17. [安全实践](#17-安全实践)
18. [CI/CD 持续集成与部署](#18-cicd-持续集成与部署)
19. [日志治理与可观测性](#19-日志治理与可观测性)
20. [gRPC 实战指南](#20-grpc-实战指南)
21. [代码规范与最佳实践](#21-代码规范与最佳实践)

---

## 1. 类型系统：继承 vs 组合

### Java 的继承

```java
// Java: 使用继承实现代码复用
public class User {
    private String id;
    private String email;
    // getters/setters
}

public class AdminUser extends User {
    private String adminLevel;
}
```

### Go 的组合（没有继承）

```go
// Go: 使用组合代替继承
type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

// AdminUser 通过组合 User 来复用
type AdminUser struct {
    User        // 嵌入 User，自动获得其字段
    AdminLevel string `json:"admin_level"`
}
```

### 嵌入 struct 的特性

```go
admin := AdminUser{
    User: User{
        ID:    "1",
        Email: "admin@example.com",
    },
    AdminLevel: "super",
}

// 可以直接访问嵌入类型的字段
fmt.Println(admin.ID)        // 继承自 User
fmt.Println(admin.Email)     // 继承自 User
fmt.Println(admin.AdminLevel) // 自己的字段
```

### 接口组合

```go
// Java: 接口可以多继承
public interface Readable { void read(); }
public interface Writeable { void write(); }
public interface ReadWriter extends Readable, Writeable {}

// Go: 接口隐式实现，组合更自然
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 组合接口
type ReadWriter interface {
    Reader
    Writer
}
```

### 最佳实践

| Java | Go |
|------|-----|
| 继承 (`extends`) | 嵌入 (`struct`) |
| implements 显式实现 | 接口隐式实现 |
| 抽象类 | 嵌入 + 接口 |

---

## 2. 分层架构：Entity/DTO/VO 的取舍

### Java Spring Boot 的标准分层

```
┌─────────────────────────────────────────┐
│  Controller (@RestController)            │  接收请求
├─────────────────────────────────────────┤
│  Service (@Service)                     │  业务逻辑
├─────────────────────────────────────────┤
│  Repository (@Repository / JPA)         │  数据访问
└─────────────────────────────────────────┘
         │
         ▼
┌─────────────┬─────────────┬─────────────┐
│   Entity    │    DTO      │     VO      │
│ (数据库表)   │ (请求/响应)  │  (视图对象)  │
└─────────────┴─────────────┴─────────────┘
```

### Go 的灵活选择

Go 没有强制的分层，但常见模式有三种：

#### 方案 A：Entity 直连（最简单）

```go
// 数据库模型
type User struct {
    ID        string    `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex"`
    Name      string
    Password  string    `json:"-"`  // 关键：用 json:"-" 隐藏敏感字段
    CreatedAt time.Time
}

// Handler 直接返回
func GetUser(c *gin.Context) {
    var user model.User
    database.DB.First(&user, id)
    response.Success(c, user)  // Password 不会被序列化
}
```

**适用**：内部系统、快速开发、API 与数据库结构一致

#### 方案 B：Request DTO + 直接返回 Entity

```go
// 请求 DTO - 验证请求参数
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Name     string `json:"name" binding:"required,min=2"`
    Password string `json:"password" binding:"required,min=6"`
}

// 响应直接用 Entity，但用 json:"-" 过滤敏感字段
type User struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Password string `json:"-"`  // 不返回
    Name     string `json:"name"`
}
```

**适用**：大多数业务系统

#### 方案 C：完整 DTO 分层

```go
// 请求 DTO
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Name     string `json:"name" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
}

// 响应 DTO
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

// 带关联的响应 DTO
type UserWithPostsResponse struct {
    UserResponse
    PostCount int    `json:"post_count"`
    LatestPost string `json:"latest_post,omitempty"`
}

// Handler
func GetUser(c *gin.Context) {
    user := service.GetUser(id)
    response.Success(c, UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt,
    })
}
```

**适用**：对外公开 API、需要版本控制、复杂响应格式

### 决策树

```
返回给前端的字段是否和 Entity 完全一致？
    │
    ├── 是 → 用 json:"-" 过滤敏感字段，直接返回 Entity
    │
    └── 否 → 需要定义 DTO
              │
              ├── 只过滤字段 → Request DTO + Entity + json:"-"
              │
              └── 字段不同/聚合 → 完整 DTO
```

### 代码组织建议

```
internal/
├── model/           # 数据库实体
│   ├── user.go
│   └── post.go
│
├── dto/             # Data Transfer Objects
│   ├── request/     # 请求 DTO
│   │   ├── user.go
│   │   └── post.go
│   └── response/    # 响应 DTO
│       ├── user.go
│       └── post.go
│
├── response/        # 统一响应封装
│   └── response.go
│
└── handler/         # HTTP 处理器
    └── user.go
```

---

## 3. 包管理：Maven/Gradle vs Go Modules

### Java (Maven/Gradle)

```xml
<!-- Maven: pom.xml -->
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-web</artifactId>
    <version>3.2.0</version>
</dependency>
```

### Go (Go Modules)

```bash
# 初始化项目
go mod init github.com/yourname/project

# 依赖自动管理在 go.mod
```

```go
// go.mod 示例
module github.com/yourname/project

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
)
```

### 依赖管理命令对比

| 操作 | Java (Maven) | Go |
|------|--------------|-----|
| 添加依赖 | 编辑 pom.xml | `go get package@version` |
| 更新依赖 | `mvn versions:use-latest` | `go get -u package` |
| 下载依赖 | `mvn install` | `go mod download` |
| 清理缓存 | `mvn clean` | `go clean -modcache` |
| 查看依赖 | `mvn dependency:tree` | `go mod graph` |

---

## 4. 依赖注入：Spring IOC vs 构造函数注入

### Java Spring 的自动注入

```java
@Service
public class UserService {
    @Autowired
    private UserRepository userRepository;  // 自动注入

    @Autowired
    private EmailService emailService;
}
```

### Go 的构造函数注入（推荐）

```go
// 显式依赖，通过构造函数注入
type UserService struct {
    repo  *UserRepository   // 具体类型
    email *EmailService    // 具体类型
}

// 构造函数
func NewUserService(repo *UserRepository, email *EmailService) *UserService {
    return &UserService{
        repo:  repo,
        email: email,
    }
}

// 在 main 或 wire/fx 中组装
func main() {
    userRepo := NewUserRepository(db)
    emailSvc := NewEmailService(smtp)
    userSvc := NewUserService(userRepo, emailSvc)

    server := NewServer(userSvc)
    server.Start()
}
```

### 接口解耦（可选但推荐）

```go
// 定义接口（类似 Java 接口）
type UserRepository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
}

// 实现
type PostgresUserRepository struct {
    db *gorm.DB
}

func (p *PostgresUserRepository) FindByID(id string) (*User, error) {
    // ...
}

// Service 依赖接口
type UserService struct {
    repo UserRepository  // 接口类型
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 依赖注入框架（可选）

| Java | Go |
|------|-----|
| Spring IoC | fx (Uber) |
| Guice | google/wire |
| CDI | dig (uber) |

**Go 哲学**：优先使用简单构造函数注入，框架是可选的。

---

## 5. 错误处理：Exception vs Error

### Java 的异常机制

```java
// Java: 抛出异常
public User findById(String id) throws UserNotFoundException {
    User user = repository.findById(id);
    if (user == null) {
        throw new UserNotFoundException("User not found: " + id);
    }
    return user;
}

// 调用方捕获
try {
    User user = service.findById("123");
} catch (UserNotFoundException e) {
    return Response.notFound();
}
```

### Go 的错误处理

```go
// Go: 返回 error
func FindByID(id string) (*User, error) {
    var user User
    if err := db.First(&user, id).Error; err != nil {
        return nil, fmt.Errorf("user not found: %s, err: %w", id, err)
    }
    return &user, nil
}

// 调用方检查
user, err := FindByID("123")
if err != nil {
    if errors.Is(err, ErrNotFound) {
        response.NotFound(c, "用户不存在")
    } else {
        response.InternalError(c, "查询失败")
    }
    return
}
```

### Go 错误 vs Java 异常的对比

| 特性 | Java Exception | Go Error |
|------|---------------|----------|
| 声明 | `throws` 关键字 | 返回值 `error` |
| 检查时机 | 编译时 (checked) | 运行时 |
| 处理方式 | try-catch | if err != nil |
| 传播 | 自动向上抛 | 手动返回 |
| 性能 | 创建异常开销大 | 轻量 |

### Go 的最佳实践

#### 1. 错误只处理一次

```go
// ❌ 错误：检查后忽略
data, _ := os.ReadFile("config.json")  // 忽略错误

// ✅ 正确：检查并处理
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("读取配置失败: %w", err)
}
```

#### 2. 使用哨兵错误（Sentinel Errors）

```go
// 定义错误
var (
    ErrNotFound     = errors.New("record not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// 使用
if errors.Is(err, ErrNotFound) {
    // 处理找不到的情况
}
```

#### 3. 错误包装

```go
// 添加上下文但不丢失原始错误
if err := db.First(&user).Error; err != nil {
    return fmt.Errorf("查询用户失败: %w", err)
}
```

#### 4. 自定义错误类型

```go
type NotFoundError struct {
    Entity string
    ID     string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Entity, e.ID)
}

// 使用
return nil, &NotFoundError{Entity: "User", ID: id}

// 检查
var notFoundErr *NotFoundError
if errors.As(err, &notFoundErr) {
    fmt.Println(notFoundErr.Entity)  // "User"
}
```

#### 5. 统一错误响应

```go
// internal/response/response.go
func Fail(c *gin.Context, code int, errMsg string) {
    c.JSON(http.StatusOK, ApiResponse{
        Code:    code,
        Message: errMsg,
        TraceID: generateTraceID(),
    })
}

// 使用
if err != nil {
    response.Fail(c, response.CodeResourceNotFound, "用户不存在")
    return
}
```

---

## 6. 并发模型：Thread vs Goroutine

### Java 线程

```java
// Java: 创建新线程
Thread thread = new Thread(() -> {
    // 耗时操作
    doHeavyWork();
});
thread.start();

// 或线程池
ExecutorService pool = Executors.newFixedThreadPool(10);
Future<String> future = pool.submit(() -> "result");
String result = future.get();  // 阻塞等待
```

### Go 的 Goroutine

```go
// Go: 启动轻量级协程
go func() {
    doHeavyWork()
}()

// 带通道的并发
ch := make(chan string)
go func() {
    result := doHeavyWork()
    ch <- result  // 发送到通道
}()

result := <-ch  // 从通道接收
```

### 核心对比

| 特性 | Java Thread | Go Goroutine |
|------|-------------|--------------|
| 内存占用 | 1-2MB/线程 | 2KB/协程 |
| 创建成本 | 高 | 极低 |
| 切换成本 | 高（内核态） | 低（用户态） |
| 最大数量 | 几千个 | 数十万个 |
| 通信方式 | 共享内存 | CSP (Channel) |

### Go 并发模式

#### 1. Channel 通信

```go
// 单向通道
func producer(ch chan<- string) {  // 只能发送
    ch <- "hello"
}

func consumer(ch <-chan string) {  // 只能接收
    msg := <-ch
    fmt.Println(msg)
}

func main() {
    ch := make(chan string)
    go producer(ch)
    go consumer(ch)
    time.Sleep(time.Second)
}
```

#### 2. WaitGroup 等待完成

```go
import "sync"

func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)  // 增加计数
        go func(id int) {
            defer wg.Done()  // 完成时减少计数
            doWork(id)
        }(i)
    }

    wg.Wait()  // 等待所有完成
}
```

#### 3. Select 多路复用

```go
select {
case msg := <-ch1:
    fmt.Println("收到 ch1:", msg)
case msg := <-ch2:
    fmt.Println("收到 ch2:", msg)
case <-time.After(time.Second):
    fmt.Println("超时")
}
```

#### 4. Context 取消

```go
func longRunningTask(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()  // 被取消
        default:
            // 执行工作
        }
    }
}

// 使用
ctx, cancel := context.WithCancel(context.Background())
go longRunningTask(ctx)
cancel()  // 取消
```

---

## 7. 泛型：Java Generics vs Go Generics

### Java 泛型

```java
// 泛型类和接口
public class Box<T> {
    private T value;
    public T get() { return value; }
    public void set(T value) { this.value = value; }
}

// 泛型方法
public <T> List<T> filter(List<T> list, Predicate<T> predicate) {
    return list.stream().filter(predicate).collect(Collectors.toList());
}

// 约束
public <T extends Comparable<T>> T max(T a, T b) {
    return a.compareTo(b) > 0 ? a : b;
}
```

### Go 1.18+ 泛型

```go
// 泛型结构体
type Box[T any] struct {
    value T
}

func (b Box[T]) Get() T {
    return b.value
}

// 泛型函数
func Filter[T any](list []T, predicate func(T) bool) []T {
    result := make([]T, 0)
    for _, item := range list {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

// 约束
type Comparable interface {
    ~int | ~float64 | ~string  // 支持的类型
    CompareTo(other Comparable) int
}

// 使用泛型
func Max[T Comparable](a, b T) T {
    if a.CompareTo(b) > 0 {
        return a
    }
    return b
}
```

### 约束对比

| Java | Go | 说明 |
|------|-----|------|
| `<T>` | `type Parameter[T any]` | 泛型参数 |
| `<T extends A>` | `T interface{A}` | 类型约束 |
| `? extends A` | `~T` (via) | 协变 |
| `? super A` | 无直接等价 | 逆变 |

---

## 8. ORM：JPA/Hibernate vs GORM

### Java JPA/Hibernate

```java
// 实体
@Entity
@Table(name = "users")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private String id;

    @Column(unique = true, nullable = false)
    private String email;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL)
    private List<Post> posts;

    @CreatedDate
    private LocalDateTime createdAt;
}

// Repository
public interface UserRepository extends JpaRepository<User, String> {
    Optional<User> findByEmail(String email);

    @Query("SELECT u FROM User u WHERE u.name LIKE %:name%")
    List<User> searchByName(@Param("name") String name);
}
```

### Go GORM

```go
// 模型
type User struct {
    ID        string    `gorm:"primaryKey;type:varchar(64)"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Name      string    `gorm:"not null"`
    Posts     []Post    `gorm:"foreignKey:UserID"`  // 一对多
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`  // 软删除
}

// 使用
var user User
result := db.First(&user, "id = ?", id)  // 查找

// 链式查询
var users []User
db.Where("name LIKE ?", "%"+name+"%").
   Order("created_at DESC").
   Limit(10).
   Find(&users)

// 关联查询
db.Preload("Posts").First(&user)
```

### GORM 常用标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `not null` | 非空 | `gorm:"not null"` |
| `type` | 指定列类型 | `gorm:"type:varchar(255)"` |
| `default` | 默认值 | `gorm:"default:0"` |
| `index` | 普通索引 | `gorm:"index"` |
| `foreignKey` | 外键 | `gorm:"foreignKey:UserID"` |
| `serializer` | 序列化 | `gorm:"serializer:json"` |

---

## 9. 接口设计：Interface 的哲学

### Java 的接口

```java
// 显式实现接口
public class UserService implements UserServiceInterface {
    @Override
    public void createUser(User user) {
        // 实现
    }
}

// 接口定义
public interface UserServiceInterface {
    void createUser(User user);
    User getUser(String id);
}
```

### Go 的隐式实现

```go
// 接口定义（一般放在使用者这边）
type UserFinder interface {
    FindByID(id string) (*User, error)
}

// 任何实现了 FindByID 方法的类型都满足接口
type PostgresUserRepo struct {}

func (p *PostgresUserRepo) FindByID(id string) (*User, error) {
    // 实现
}

// 无需显式声明 implements
var _ UserFinder = (*PostgresUserRepo)(nil)  // 编译时检查是否实现接口
```

### Go 接口设计原则

#### 1. 接口越小越好

```go
// ❌ 大接口（冗余）
type UserService interface {
    Create(user *User) error
    Get(id string) (*User, error)
    Update(user *User) error
    Delete(id string) error
    List() ([]User, error)
    FindByEmail(email string) (*User, error)
    FindByName(name string) ([]User, error)
}

// ✅ 小接口（单一职责）
type Creator interface { Create(any) any }
type Finder interface { Find(any) any }
type Deleter interface { Delete(any) any }
```

#### 2. 接口放在使用者包中

```go
// user/service.go (使用者)
type UserService interface {
    GetUser(id string) (*User, error)
}

// user/postgres.go (实现者)
// 不需要声明 implements，直接实现方法即可
```

#### 3. 空接口 `any` (interface{})

```go
// 相当于 Java 的 Object
var anything interface{} = "hello"
anything = 123
anything = []int{1, 2, 3}

// 类型断言
str, ok := anything.(string)
if ok {
    fmt.Println(str)
}

// switch 判断类型
switch v := anything.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
}
```

---

## 10. 实战建议

### 给 Java 开发者的 Go 最佳实践

#### 1. 包结构

```
// Java
com.example.project
├── controller/
├── service/
├── repository/
├── entity/
└── dto/

// Go (简洁扁平)
internal/
├── handler/    // Controller
├── service/
├── repository/
├── model/      // Entity + DTO
└── response/
```

#### 2. 命名约定

| Java | Go | 说明 |
|------|-----|------|
| `UserService` | `UserService` | 保持一致 |
| `userRepository` | `userRepo` | Go 偏好简短 |
| `getUserById()` | `GetUserByID()` | 首字母大写即导出 |
| `MAX_SIZE` | `MaxSize` | Go 偏好驼峰 |

#### 3. 错误处理习惯

```go
// Go 的错误检查
if err != nil {
    return fmt.Errorf("操作: %w", err)
}

// 不要忽略错误
data, _ := os.ReadFile("...")  // ❌
data, err := os.ReadFile("...") // ✅
```

#### 4. 初始化习惯

```go
// 声明 + 初始化
var users []User        // 声明 nil slice
users := []User{}       // 声明空 slice
users := make([]User, 0) // 预分配

var m map[string]User   // 声明 nil map
m := make(map[string]User) // 初始化
```

#### 5. 值 vs 指针

```go
// 小结构/基本类型 → 值传递
func add(a, b int) int { return a + b }

// 大结构/需要修改 → 指针传递
func (u *User) Update() error { /* 修改 */ }

// 数据库模型通常用指针
type User struct { ... }
var user *User  // 常见
```

#### 6. 常用工具库

| 用途 | Java | Go |
|------|------|-----|
| Web 框架 | Spring Boot | Gin / Echo / Chi |
| ORM | JPA / Hibernate | GORM / sqlx |
| 依赖注入 | Spring | fx / wire / 手动 |
| 日志 | Logback / SLF4J | slog / zap / logrus |
| 验证 | Hibernate Validator | go-playground/validator |
| JSON | Jackson | encoding/json / json-iterator |
| HTTP Client | RestTemplate / WebClient | net/http / resty |
| 配置文件 | YAML / Properties | TOML / YAML / 环境变量 |

---

## 11. 测试实践：JUnit vs testing

### Java JUnit

```java
// Java: JUnit 5
@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository userRepository;

    @InjectMock
    private UserService userService;

    @Test
    void testFindById() {
        // given
        User mockUser = new User("1", "test@example.com");
        when(userRepository.findById("1")).thenReturn(mockUser);

        // when
        User result = userService.findById("1");

        // then
        assertEquals("1", result.getId());
        assertEquals("test@example.com", result.getEmail());
    }

    @Test
    void testFindById_NotFound() {
        when(userRepository.findById("999")).thenReturn(null);

        assertThrows(UserNotFoundException.class,
            () -> userService.findById("999"));
    }
}
```

### Go 标准测试

```go
// user_service_test.go
package service_test  // 测试包后缀 _test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestXxx 开头，自动被测试运行
func TestUserService_FindByID(t *testing.T) {
    // 准备 mock（Go 常用 sqlmock, gomock 等库）
    repo := NewMockUserRepo()
    svc := NewUserService(repo)

    // 设置 mock 行为
    repo.On("FindByID", "1").Return(&User{ID: "1", Email: "test@example.com"}, nil)

    // 执行
    user, err := svc.FindByID("1")

    // 断言
    assert.NoError(t, err)
    assert.Equal(t, "1", user.ID)
    assert.Equal(t, "test@example.com", user.Email)
}

func TestUserService_FindByID_NotFound(t *testing.T) {
    repo := NewMockUserRepo()
    svc := NewUserService(repo)

    repo.On("FindByID", "999").Return(nil, ErrNotFound)

    _, err := svc.FindByID("999")
    assert.ErrorIs(t, err, ErrNotFound)
}
```

### Go 常用测试工具

| 功能 | Java | Go |
|------|------|-----|
| Mock 框架 | Mockito, EasyMock | testify/mock, gomock |
| 断言库 | AssertJ, Hamcrest | testify/assert, go-cmp |
| 数据库测试 | Testcontainers | sqlmock, dockertest |
| HTTP 测试 | MockMvc, RestAssured | httptest, net/http/httptest |
| 覆盖率 | JaCoCo | go test -cover |

### Go 表驱动测试

```go
// Go 特色：用表驱动测试减少重复
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"正数相加", 1, 2, 3},
        {"负数相加", -1, -2, -3},
        {"正负相加", 1, -1, 0},
        {"零相加", 0, 5, 5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Go Benchmark 性能测试

```go
// 性能测试函数
func BenchmarkStringBuilder(b *testing.B) {
    var str string
    for i := 0; i < b.N; i++ {
        str = "hello" + "world"
    }
}

// 运行: go test -bench=BenchmarkStringBuilder -benchmem
```

### HTTP Handler 测试

```go
// Go 测试 HTTP Handler
func TestGetUser(t *testing.T) {
    // 创建测试路由
    r := gin.New()
    r.GET("/users/:id", GetUser)

    // 创建请求
    req, _ := http.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()

    // 执行
    r.ServeHTTP(w, req)

    // 断言
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "user")
}
```

### 测试金字塔

```
         /\
        /  \       E2E 测试 (少量)
       /____\
      /      \     集成测试 (适量)
     /________\
    /          \   单元测试 (大量)
   /____________\
```

| 层级 | Java | Go |
|------|------|-----|
| 单元测试 | JUnit + Mockito | testing + testify/mock |
| 集成测试 | SpringBootTest | dockertest / sqlmock |
| E2E 测试 | Selenium, Cypress | Selenium, playwright |

### Mock 接口的最佳实践

```go
// 定义接口（便于测试时替换）
type UserRepository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
}

// 具体实现
type PostgresUserRepo struct {
    db *gorm.DB
}

func (p *PostgresUserRepo) FindByID(id string) (*User, error) {
    var user User
    if err := p.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

// 测试时用 mock
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) FindByID(id string) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Service 依赖接口
type UserService struct {
    repo UserRepository  // 接口类型
}

// 测试
func TestUserService(t *testing.T) {
    mockRepo := new(MockUserRepo)
    svc := NewUserService(mockRepo)

    mockRepo.On("FindByID", "1").Return(&User{ID: "1"}, nil)

    user, err := svc.GetUser("1")
    assert.NoError(t, err)
    assert.Equal(t, "1", user.ID)
    mockRepo.AssertExpectations(t)
}
```

### 运行测试命令对比

| 操作 | Java (Maven) | Go |
|------|--------------|-----|
| 运行所有测试 | `mvn test` | `go test ./...` |
| 运行特定测试 | `mvn test -Dtest=UserTest` | `go test -run TestUser` |
| 显示覆盖 | `mvn test -jacoco` | `go test -coverprofile=c.out` |
| Benchmark | - | `go test -bench=. -benchmem` |
| 并行测试 | `@Parallel` | `-parallel 4` |

---

## 12. 缓存设计模式：Redis 实践

### Java Spring Cache

```java
// Java: Spring Cache 注解
@Service
public class UserService {

    @Cacheable(value = "users", key = "#id")
    public User findById(String id) {
        return userRepository.findById(id);
    }

    @CacheEvict(value = "users", key = "#id")
    public void deleteById(String id) {
        userRepository.deleteById(id);
    }

    @CachePut(value = "users", key = "#user.id")
    public User update(User user) {
        return userRepository.save(user);
    }
}
```

### Go + Redis 实现

#### 1. 手动缓存模式

```go
type UserService struct {
    repo  *UserRepo
    cache *redis.Client
}

func (s *UserService) GetUser(id string) (*User, error) {
    ctx := context.Background()
    cacheKey := fmt.Sprintf("user:%s", id)

    // 1. 先查缓存
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. 缓存不存在，查数据库
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // 3. 写入缓存（设置过期时间）
    data, _ := json.Marshal(user)
    s.cache.Set(ctx, cacheKey, data, 30*time.Minute)

    return user, nil
}

func (s *UserService) DeleteUser(id string) error {
    ctx := context.Background()

    // 1. 删除数据库
    if err := s.repo.Delete(id); err != nil {
        return err
    }

    // 2. 删除缓存
    cacheKey := fmt.Sprintf("user:%s", id)
    s.cache.Del(ctx, cacheKey)

    return nil
}
```

#### 2. 缓存辅助函数

```go
// cache/cache.go
package cache

type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

// GetOrSet 获取缓存，不存在则调用 factory 并缓存
func (c *Cache) GetOrSet(ctx context.Context, key string, factory func() (any, error)) (any, error) {
    // 尝试获取缓存
    val, err := c.client.Get(ctx, key).Result()
    if err == nil {
        return json.Unmarshal([]byte(val))
    }

    // 缓存未命中，调用 factory
    result, err := factory()
    if err != nil {
        return nil, err
    }

    // 存入缓存
    data, _ := json.Marshal(result)
    c.client.Set(ctx, key, data, c.ttl)

    return result, nil
}

// Invalidate 删除缓存
func (c *Cache) Invalidate(ctx context.Context, keys ...string) error {
    return c.client.Del(ctx, keys...).Err()
}
```

#### 3. 缓存模式实践

```go
// ========== Cache-Aside (旁路缓存) ==========
// 最常用：先缓存，缓存未命中查 DB 并回填

func GetUser(service *UserService, id string) (*User, error) {
    // 1. 查缓存
    user, err := service.cache.GetUser(id)
    if err == nil {
        return user, nil  // 命中
    }

    // 2. 缓存未命中，查 DB
    user, err = service.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // 3. 回填缓存
    service.cache.SetUser(user)

    return user, nil
}

// ========== Write-Through (穿透写) ==========
// 写入时同步更新缓存

func CreateUser(service *UserService, user *User) error {
    // 1. 写入 DB
    if err := service.repo.Create(user); err != nil {
        return err
    }

    // 2. 同步写入缓存
    service.cache.SetUser(user)

    return nil
}

// ========== Write-Behind (回写) ==========
// 写入时只更新缓存，异步批量写 DB（高并发场景）

func UpdateUser(service *UserService, user *User) error {
    // 1. 只更新缓存
    service.cache.SetUser(user)

    // 2. 异步写 DB（可通过消息队列）
    go func() {
        service.repo.Update(user)
    }()

    return nil
}
```

### 缓存失效策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| TTL 过期 | 设置固定过期时间 | 大多数场景 |
| LRU | 最近最少使用淘汰 | 缓存空间有限 |
| Write-Invalidate | 写入时删除缓存 | 数据一致性要求高 |
| Read-Invalidation | 读取时检查版本 | 强一致性 |

### Redis 常见数据结构应用

| 数据结构 | Java | Go | 适用场景 |
|----------|------|-----|----------|
| String | `stringRedisTemplate.opsForValue()` | `client.Get/Set` | 简单值、序列化对象 |
| Hash | `opsForHash()` | `HGetAll/HSet` | 对象字段 |
| List | `opsForList()` | `LPush/RPop` | 队列、最新N条 |
| Set | `opsForSet()` | `SAdd/SMembers` | 标签、去重 |
| Sorted Set | `opsForZSet()` | `ZAdd/ZRange` | 排行榜、延时队列 |

### 分布式锁

```go
// Redis 分布式锁实现
import "github.com/redis/go-redis/v9"

func AcquireLock(client *redis.Client, key string, ttl time.Duration) (bool, error) {
    ctx := context.Background()
    // SET key value NX EX seconds
    result, err := client.SetNX(ctx, key, "1", ttl).Result()
    return result, err
}

func ReleaseLock(client *redis.Client, key string) error {
    ctx := context.Background()
    return client.Del(ctx, key).Err()
}

// 使用
func UpdateUserTx(service *UserService, id string, update func(*User)) error {
    lockKey := fmt.Sprintf("lock:user:%s", id)

    // 获取锁（10秒超时）
    acquired, _ := AcquireLock(service.redis, lockKey, 10*time.Second)
    if !acquired {
        return errors.New("系统繁忙，请稍后重试")
    }
    defer ReleaseLock(service.redis, lockKey)

    // 业务操作
    user, err := service.GetUser(id)
    if err != nil {
        return err
    }
    update(user)
    return service.repo.Update(user)
}
```

### 缓存雪崩、穿透、击穿

```go
// ========== 雪崩：大量 key 同时过期 ==========
// 解决方案：过期时间 + 随机偏移
ttl := 30 * time.Minute + time.Duration(rand.Intn(5))*time.Minute
cache.Set(ctx, key, data, ttl)

// 或用不过期 + 版本号
func GetUserWithVersion(service *UserService, id string) (*User, error) {
    // 查询时带上版本号，版本变化说明数据更新了
}

// ========== 穿透：查询不存在的数据 ==========
// 解决方案：布隆过滤器 / 空值缓存
if user == nil {
    // 缓存空值，短过期时间
    cache.Set(ctx, fmt.Sprintf("empty:%s", id), "1", 1*time.Minute)
}

// ========== 击穿：热点 key 过期瞬间大量请求 ==========
// 解决方案：互斥锁 / 单飞请求
var user *User
var err error

// 单飞请求（只有一个协程去查 DB）
user, err = service.singleflight.Do(id, func() (any, error) {
    return service.repo.FindByID(id)
})
```

---

## 13. 微服务架构：Spring Cloud vs Go 微服务

### Java Spring Cloud 组件

```
┌─────────────────────────────────────────────────────────────┐
│                        Spring Cloud                          │
├─────────────────────────────────────────────────────────────┤
│  服务发现     │  Eureka / Nacos                              │
│  网关         │  Gateway / Zuul                              │
│  配置中心     │  Config Server / Nacos                       │
│  负载均衡     │  Ribbon / LoadBalancer                       │
│  熔断器       │  Hystrix / Resilience4j                     │
│  分布式事务   │  Seata                                       │
│  消息队列     │  Kafka / RabbitMQ                            │
│  链路追踪     │  Sleuth + Zipkin / Jaeger                   │
└─────────────────────────────────────────────────────────────┘
```

### Go 微服务生态

```
┌─────────────────────────────────────────────────────────────┐
│                        Go 微服务生态                         │
├─────────────────────────────────────────────────────────────┤
│  服务发现     │  Consul / etcd / Kubernetes DNS             │
│  网关         │  Kong / Caddy / Traefik / 自己实现           │
│  配置中心     │  etcd / Consul / Viper                       │
│  负载均衡     │  Kubernetes Service / Envoy                 │
│  熔断器       │  sony/gobreaker / 第三方库                  │
│  分布式事务   │  DTM / 自己实现 Saga                         │
│  消息队列     │  Kafka / RabbitMQ / NATS / Redis Stream     │
│  链路追踪     │  OpenTelemetry + Jaeger / Zipkin             │
└─────────────────────────────────────────────────────────────┘
```

### 服务注册与发现

#### Java (Nacos)

```java
// 服务提供者
@SpringBootApplication
@EnableDiscoveryClient
public class UserServiceApplication {
    public static void main(String[] args) {
        SpringApplication.run(UserServiceApplication.class, args);
    }
}

// application.yml
spring:
  cloud:
    nacos:
      discovery:
        server-addr: nacos-server:8848
```

#### Go (Consul)

```go
// 服务注册
import "github.com/hashicorp/consul/api"

func RegisterService() {
    config := api.DefaultConfig()
    config.Address = "consul:8500"
    client, _ := api.NewClient(config)

    // 注册
    registration := &api.AgentServiceRegistration{
        ID:   "user-service-1",
        Name: "user-service",
        Port: 8080,
        Check: &api.AgentCheck{
            HTTP:     "http://localhost:8080/health",
            Interval: "10s",
        },
    }
    client.Agent().ServiceRegister(registration)
}

// 服务发现
services, _, _ := client.Health().Service("user-service", "", true, nil)
for _, svc := range services {
    fmt.Println(svc.Service.Address, svc.Service.Port)
}
```

### API 网关

#### Java (Spring Cloud Gateway)

```java
@Configuration
public class GatewayConfig {
    @Bean
    public RouteLocator customRouteLocator(RouteLocatorBuilder builder) {
        return builder.routes()
            .route("user-service", r -> r
                .path("/api/users/**")
                .uri("lb://user-service"))
            .route("order-service", r -> r
                .path("/api/orders/**")
                .uri("lb://order-service"))
            .build();
    }
}
```

#### Go (Kong / 自定义)

```go
// Kong declarative config (kong.yml)
services:
  - name: user-service
    url: http://user-service:8080
    routes:
      - name: user-route
        paths:
          - /api/users
plugins:
  - name: rate-limiting
    config:
      minute: 100

// 或自己实现简单网关
type Gateway struct {
    routes map[string]string
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    if target, ok := g.routes[path]; ok {
        http.Redirect(w, r, target, http.StatusFound)
        return
    }
    http.NotFound(w, r)
}
```

### 熔断器

#### Java (Resilience4j)

```java
@CircuitBreaker(name = "userService", fallbackMethod = "fallback")
public User getUser(String id) {
    return userClient.findById(id);
}

public User fallback(String id, Throwable t) {
    return User.defaultUser();  // 降级返回
}
```

#### Go (gobreaker)

```go
import "github.com/sony/gobreaker"

var cb *gobreaker.CircuitBreaker

func init() {
    cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "user-service",
        MaxRequests: 3,        // 熔断开放后，最小请求数
        Interval:    10 * time.Second,  // 熔断关闭后，清零计数周期
        Timeout:     30 * time.Second, // 熔断开放持续时间
    })
}

func GetUser(id string) (*User, error) {
    result, err := cb.Execute(func() (any, error) {
        return userClient.FindByID(id)
    })

    if errors.Is(err, gobreaker.ErrOpenState) {
        return defaultUser(), nil  // 降级
    }
    return result.(*User), err
}
```

### 配置中心

#### Java (Spring Cloud Config + Nacos)

```java
// application.yml
spring:
  cloud:
    config:
      server:
        git:
          uri: https://github.com/your-org/config-repo
          default-label: main
```

#### Go (Viper + etcd)

```go
import "github.com/spf13/viper"
import "go.etcd.io/etcd/client/v3"

func LoadConfig() (*viper.Viper, error) {
    v := viper.New()

    // 从 etcd 读取
    cli, _ := client.New(client.Config{
        Endpoints: []string{"etcd:2379"},
    })
    defer cli.Close()

    resp, _ := cli.Get(context.Background(), "/app/config")
    for _, kv := range resp.Kvs {
        v.Set(string(kv.Key), string(kv.Value))
    }

    // 或从文件读取
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AutomaticEnv()

    return v, nil
}
```

### 分布式事务

#### Java (Seata)

```java
@GlobalTransactional
public void createOrder(Order order) {
    // AT 模式自动处理
    orderService.create(order);
    inventoryService.deduct(order.getProductId(), order.getQuantity());
}
```

#### Go (DTM)

```go
import "github.com/dtmf/dtmcli"

func CreateOrder(saga *dtmcli.Saga) error {
    saga.Add(
        "http://inventory-service/deduct",
        "http://inventory-service/compensate",
        map[string]string{"product_id": "P001", "quantity": "1"},
    )
    saga.Add(
        "http://order-service/create",
        "http://order-service/compensate",
        map[string]string{"order_id": "O001"},
    )

    return saga.Submit()
}
```

### 链路追踪

#### Java (OpenTelemetry + Jaeger)

```java
// Spring Boot 3.x 内置支持
@SpringBootApplication
@EnableTracing
public class UserService {
    public static void main(String[] args) {
        SpringApplication.run(UserService.class, args);
    }
}
```

#### Go (OpenTelemetry)

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer() *trace.TracerProvider {
    exporter, _ := jaeger.New(
        jaeger.WithCollectorEndpoint("http://jaeger:14268/api/traces"),
    )

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
    )

    otel.SetTracerProvider(tp)
    return tp
}

// 使用
func GetUser(tracer trace.Tracer, repo *UserRepo) (*User, error) {
    ctx, span := tracer.Start(context.Background(), "GetUser")
    defer span.End()

    span.SetAttributes("user.id", "123")

    user, err := repo.FindByID(ctx, "123")
    if err != nil {
        span.RecordError(err)
    }

    return user, err
}
```

### gRPC vs REST

| 特性 | REST | gRPC |
|------|------|------|
| 协议 | HTTP/1.1 | HTTP/2 |
| 数据格式 | JSON | Protocol Buffers |
| 代码生成 | OpenAPI/Swagger | protoc |
| 流式支持 | SSE/Long Polling | 原生双向流 |
| 适用场景 | 外部 API、Browser | 微服务内部 |

#### Go gRPC 示例

```go
// 定义 proto
// user.proto
syntax = "proto3";
package user;

service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc ListUsers(ListUsersRequest) returns (stream User);  // 服务端流
}

message GetUserRequest {
    string id = 1;
}

// 生成代码
// protoc --go_out=. --go-grpc_out=. user.proto

// 服务端
type UserServer struct {
    proto.UnimplementedUserServiceServer
}

func (s *UserServer) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.User, error) {
    return &proto.User{
        Id:    req.Id,
        Name:  "Test User",
        Email: "test@example.com",
    }, nil
}

// 客户端
conn, _ := grpc.Dial("user-service:8080", grpc.WithInsecure())
client := proto.NewUserServiceClient(conn)
user, _ := client.GetUser(ctx, &proto.GetUserRequest{Id: "1"})
```

### 微服务健康检查

```go
// 健康检查接口
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    checks := map[string]bool{
        "database": checkDB(),
        "redis":    checkRedis(),
        "external": checkExternal(),
    }

    allHealthy := true
    for _, ok := range checks {
        if !ok {
            allHealthy = false
            break
        }
    }

    status := map[string]any{
        "status": "healthy",
        "checks": checks,
    }

    if !allHealthy {
        status["status"] = "degraded"
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(status)
}

// Kubernetes probes
// livenessProbe:
//   httpGet:
//     path: /health
//     port: 8080
//   initialDelaySeconds: 10
//   periodSeconds: 5
// readinessProbe:
//   httpGet:
//     path: /ready
//     port: 8080
```

---

## 14. Docker 与 Kubernetes 部署

### Java Dockerfile

```dockerfile
# Maven 多阶段构建
FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=builder /app/target/myapp.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### Go Dockerfile

```dockerfile
# Go 多阶段构建（推荐）
# 第一阶段：构建
FROM golang:1.21-alpine AS builder
WORKDIR /build

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制代码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：运行
FROM alpine:3.19
WORKDIR /app

# 安装 CA 证书（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /build/main .
COPY --from=builder /build/config ./config

# 创建非 root 用户
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080
ENTRYPOINT ["./main"]
```

### Docker Compose

#### Java Spring Boot + MySQL

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      SPRING_DATASOURCE_URL: jdbc:mysql://mysql:3306/mydb
      SPRING_DATASOURCE_USERNAME: root
      SPRING_DATASOURCE_PASSWORD: secret
    depends_on:
      mysql:
        condition: service_healthy

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mydb
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

#### Go + MySQL + Redis

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: root:secret@tcp(mysql:3306)/mydb?charset=utf8mb4
      REDIS_ADDR: redis:6379
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mydb
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

### Kubernetes 部署

#### Java Spring Boot Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
        - name: user-service
          image: myrepo/user-service:v1.0.0
          ports:
            - containerPort: 8080
          env:
            - name: SPRING_PROFILES_ACTIVE
              value: "k8s"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
          resources:
            requests:
              memory: "512Mi"
              cpu: "250m"
            limits:
              memory: "1Gi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /actuator/health/liveness
              port: 8080
            initialDelaySeconds: 60
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /actuator/health/readiness
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 5
```

#### Go Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
        - name: user-service
          image: myrepo/user-service:v1.0.0
          ports:
            - containerPort: 8080
          env:
            - name: ENV
              value: "production"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

#### Kubernetes Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP  # 或 LoadBalancer / NodePort
```

#### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  app.yaml: |
    database:
      host: mysql
      port: 3306
      name: mydb
    cache:
      ttl: 30m
    log:
      level: info
---
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
stringData:
  url: "root:password@tcp(mysql:3306)/mydb?charset=utf8mb4"
```

### Helm Chart 对比

#### Java Spring Boot Helm

```yaml
# Chart.yaml
apiVersion: v2
name: user-service
version: 1.0.0
appVersion: "1.0.0"

# values.yaml
replicaCount: 3

image:
  repository: myrepo/user-service
  tag: "v1.0.0"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 8080

resources:
  requests:
    memory: 512Mi
    cpu: 250m
  limits:
    memory: 1Gi
    cpu: 500m

env:
  - name: SPRING_PROFILES_ACTIVE
    value: "prod"
```

#### Go Helm

```yaml
# Chart.yaml
apiVersion: v2
name: user-service
version: 1.0.0
appVersion: "1.0.0"

# values.yaml
replicaCount: 3

image:
  repository: myrepo/user-service
  tag: "v1.0.0"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 8080

resources:
  requests:
    memory: 128Mi
    cpu: 100m
  limits:
    memory: 256Mi
    cpu: 200m

env:
  - name: ENV
    value: "production"
```

### 资源对比

| 资源 | Java (Spring Boot) | Go |
|------|---------------------|-----|
| 内存 (JVM) | 512MB - 2GB | 64MB - 256MB |
| 启动时间 | 10-30 秒 | < 1 秒 |
| 镜像大小 | 200-500MB | 10-50MB |
| CPU | 相对较高 | 较低 |

---

## 15. Go 设计模式

Go 没有继承，所以经典 GoF 设计模式需要用组合、接口、函数来表达。

### 单例模式

#### Java

```java
public class Singleton {
    private static Singleton instance;
    public static synchronized Singleton getInstance() {
        if (instance == null) instance = new Singleton();
        return instance;
    }
}
```

#### Go

```go
// 方式 1：sync.Once（推荐）
type Config struct {
    data map[string]string
}

var (
    instance *Config
    once     sync.Once
)

func GetConfig() *Config {
    once.Do(func() {
        instance = &Config{data: make(map[string]string)}
    })
    return instance
}

// 方式 2：包级别变量（更 Go 风格）
var cfg = &Config{data: make(map[string]string)}
func GetConfig() *Config { return cfg }
```

### 工厂模式

```go
// 定义产品接口
type Storage interface {
    Save(key string, data []byte) error
    Load(key string) ([]byte, error)
}

// 具体产品
type S3Storage struct{ bucket string }
type LocalStorage struct{ path string }

func (s *S3Storage) Save(key string, data []byte) error { /* ... */ return nil }
func (s *S3Storage) Load(key string) ([]byte, error) { return nil, nil }

func (s *LocalStorage) Save(key string, data []byte) error { /* ... */ return nil }
func (s *LocalStorage) Load(key string) ([]byte, error) { return nil, nil }

// 工厂函数
func NewStorage(typ string) Storage {
    switch typ {
    case "s3":
        return &S3Storage{bucket: "mybucket"}
    case "local":
        return &LocalStorage{path: "./data"}
    default:
        return nil
    }
}
```

### 策略模式

```go
// 支付策略接口
type PaymentStrategy interface {
    Pay(amount float64) error
}

// 具体策略
type CreditCardPay struct{ cardNumber string }
type AlipayPay struct{}

func (p *CreditCardPay) Pay(amount float64) error {
    fmt.Printf("Paying %.2f with CreditCard %s\n", amount, p.cardNumber)
    return nil
}

func (p *AlipayPay) Pay(amount float64) error {
    fmt.Printf("Paying %.2f with Alipay\n", amount)
    return nil
}

// 上下文
type Payment struct {
    strategy PaymentStrategy
}

func (p *Payment) SetStrategy(s PaymentStrategy) {
    p.strategy = s
}

func (p *Payment) Checkout(amount float64) error {
    return p.strategy.Pay(amount)
}

// 使用
payment := &Payment{}
payment.SetStrategy(&CreditCardPay{cardNumber: "1234"})
payment.Checkout(100.00)
```

### 装饰器模式

```go
// 基础接口
type Handler interface {
    Handle(ctx context.Context, req *Request) (*Response, error)
}

// 基础实现
type AuthHandler struct{}

func (h *AuthHandler) Handle(ctx context.Context, req *Request) (*Response, error) {
    // 业务逻辑
    return &Response{Data: "OK"}, nil
}

// 装饰器：日志
type LoggingDecorator struct {
    next Handler
}

func (d *LoggingDecorator) Handle(ctx context.Context, req *Request) (*Response, error) {
    log.Printf("Request: %v", req)
    resp, err := d.next.Handle(ctx, req)
    log.Printf("Response: %v, Error: %v", resp, err)
    return resp, err
}

// 装饰器：缓存
type CacheDecorator struct {
    next  Handler
    cache *redis.Client
}

func (d *CacheDecorator) Handle(ctx context.Context, req *Request) (*Response, error) {
    key := fmt.Sprintf("cache:%s", req.Key)
    if data, err := d.cache.Get(ctx, key).Bytes(); err == nil {
        var resp Response
        json.Unmarshal(data, &resp)
        return &resp, nil
    }
    resp, err := d.next.Handle(ctx, req)
    if err == nil {
        data, _ := json.Marshal(resp)
        d.cache.Set(ctx, key, data, time.Minute)
    }
    return resp, err
}

// 链式调用
handler := &AuthHandler{}
handler = &LoggingDecorator{next: handler}
handler = &CacheDecorator{next: handler, cache: redisClient}
```

### 观察者模式

```go
// 事件
type Event struct {
    Type string
    Data any
}

// 观察者接口
type Observer interface {
    OnNotify(event Event)
}

// 发布者
type Publisher struct {
    observers []Observer
    mu        sync.RWMutex
}

func (p *Publisher) Subscribe(o Observer) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.observers = append(p.observers, o)
}

func (p *Publisher) Unsubscribe(o Observer) {
    p.mu.Lock()
    defer p.mu.Unlock()
    for i, obs := range p.observers {
        if obs == o {
            p.observers = append(p.observers[:i], p.observers[i+1:]...)
            break
        }
    }
}

func (p *Publisher) Notify(event Event) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    for _, o := range p.observers {
        o.OnNotify(event)
    }
}

// 具体观察者
type EmailNotifier struct{}

func (e *EmailNotifier) OnNotify(event Event) {
    if event.Type == "user.created" {
        fmt.Println("Sending welcome email...")
    }
}
```

### 生成器模式

```go
// Builder 接口
type Builder interface {
    SetName(name string) Builder
    SetAge(age int) Builder
    SetEmail(email string) Builder
    Build() *User
}

// 具体 Builder
type UserBuilder struct {
    user User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{}
}

func (b *UserBuilder) SetName(name string) Builder {
    b.user.Name = name
    return b
}

func (b *UserBuilder) SetAge(age int) Builder {
    b.user.Age = age
    return b
}

func (b *UserBuilder) SetEmail(email string) Builder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) Build() *User {
    return &b.user
}

// 使用（链式调用）
user := NewUserBuilder().
    SetName("张三").
    SetAge(30).
    SetEmail("zhangsan@example.com").
    Build()
```

### 命令模式

```go
// 命令接口
type Command interface {
    Execute() error
}

// 具体命令
type CreateUserCmd struct {
    repo *UserRepo
    user *User
}

func (c *CreateUserCmd) Execute() error {
    return c.repo.Create(c.user)
}

type DeleteUserCmd struct {
    repo *UserRepo
    userID string
}

func (c *DeleteUserCmd) Execute() error {
    return c.repo.Delete(c.userID)
}

// 调用者
type CommandQueue struct {
    commands []Command
}

func (q *CommandQueue) Add(cmd Command) {
    q.commands = append(q.commands, cmd)
}

func (q *CommandQueue) ExecuteAll() error {
    for _, cmd := range q.commands {
        if err := cmd.Execute(); err != nil {
            return err
        }
    }
    return nil
}
```

### Go 特有的模式

#### Options 模式（替代 Builder 简化版）

```go
type Server struct {
    host string
    port int
    timeout time.Duration
}

// Option 函数类型
type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) { s.timeout = timeout }
}

// 构造函数
func NewServer(opts ...Option) *Server {
    s := &Server{
        host:    "localhost",
        port:    8080,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// 使用
server := NewServer(
    WithHost("0.0.0.0"),
    WithPort(9090),
)
```

#### Functional Options Pattern 进阶

```go
type HTTPClient struct {
    timeout    time.Duration
    maxRetries int
    headers    map[string]string
}

type HTTPOption func(*HTTPClient)

func WithTimeout(t time.Duration) HTTPOption {
    return func(c *HTTPClient) { c.timeout = t }
}

func WithMaxRetries(n int) HTTPOption {
    return func(c *HTTPClient) { c.maxRetries = n }
}

func WithHeader(k, v string) HTTPOption {
    return func(c *HTTPClient) {
        if c.headers == nil {
            c.headers = make(map[string]string)
        }
        c.headers[k] = v
    }
}

func NewHTTPClient(opts ...HTTPOption) *HTTPClient {
    c := &HTTPClient{
        timeout:    30 * time.Second,
        maxRetries: 3,
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

---

## 16. 性能优化技巧

### 内存分配优化

#### 1. 对象池 (sync.Pool)

```go
// ❌ 频繁分配
func Process(data []byte) error {
    decoder := json.NewDecoder(bytes.NewReader(data))  // 每次创建
    var req Request
    decoder.Decode(&req)
    // decoder 被 GC
    return nil
}

// ✅ 使用对象池
var decoderPool = sync.Pool{
    New: func() any {
        return &json.Decoder{}
    },
}

func Process(data []byte) error {
    decoder := decoderPool.Get().(*json.Decoder)
    defer decoderPool.Put(decoder)

    decoder.Reset(bytes.NewReader(data))
    var req Request
    decoder.Decode(&req)
    return nil
}
```

#### 2. 预分配 Slice 和 Map

```go
// ❌ 动态增长
var items []Item
for _, id := range ids {
    item := fetchItem(id)
    items = append(items, item)  // 可能多次扩容
}

// ✅ 预分配容量
items := make([]Item, 0, len(ids))  // 预分配 len(ids) 容量
for _, id := range ids {
    items = append(items, fetchItem(id))
}

// ❌ 动态增长的 Map
m := make(map[string]User)
for _, id := range ids {
    m[id] = fetchUser(id)  // 可能多次扩容
}

// ✅ 预估容量
m := make(map[string]User, len(ids))
```

### 字符串优化

```go
// ❌ 字符串拼接（大量操作时效率低）
result := ""
for _, s := range strs {
    result += s  // 每次创建新字符串
}

// ✅ strings.Builder
var builder strings.Builder
for _, s := range strs {
    builder.WriteString(s)
}
result := builder.String()

// ❌ 频繁字符串转换
data := string(bytes[:])  // 复制
str := strconv.Itoa(num)   // 分配

// ✅ 使用 string([]byte) 直接转换（零拷贝，只在需要复制时复制）
data := string(bytes)  // 编译器优化
```

### JSON 序列化优化

```go
// ❌ 标准库 json（慢但通用）
data, _ := json.Marshal(user)

// ✅ json-iterator（快 3-5 倍）
import "github.com/json-iterator/go"
var json = jsoniter.ConfigCompatibleWithStandardLibrary

data, _ := json.Marshal(user)

// ✅ 使用 gogs/sjson（只修改部分字段）
import "github.com/gogs/sjson"
path := "user.name"
newData, _ := sjson.Set(data, path, "new name")
```

### 并发优化

#### 1. 协程池 (Worker Pool)

```go
// ❌ 无限制创建协程
for _, task := range tasks {
    go process(task)  // 可能创建数万个协程
}

// ✅ 协程池
func WorkerPool(workers int, jobs <-chan Job, results chan<- Result) {
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }
    wg.Wait()
}

// 使用
jobs := make(chan Job, len(tasks))
results := make(chan Result, len(tasks))

for _, task := range tasks {
    jobs <- task
}
close(jobs)

WorkerPool(10, jobs, results)
```

#### 2. errgroup 并发处理

```go
import "golang.org/x/sync/errgroup"

func FetchAll(urls []string) ([]byte, error) {
    g, ctx := errgroup.WithContext(context.Background())

    results := make([][]byte, len(urls))

    for i, url := range urls {
        i, url := i, url  // 捕获变量
        g.Go(func() error {
            resp, err := http.Get(url)
            if err != nil {
                return err
            }
            defer resp.Body.Close()

            data, err := io.ReadAll(resp.Body)
            if err != nil {
                return err
            }
            results[i] = data
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return json.Marshal(results)
}
```

### 数据库优化

#### 1. 连接池配置

```go
sqlDB, _ := db.DB()

// 设置最大连接数
sqlDB.SetMaxOpenConns(25)

// 设置最大空闲连接数
sqlDB.SetMaxIdleConns(10)

// 设置连接最大生命周期
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

#### 2. GORM 优化

```go
// ❌ 全字段查询
var users []User
db.Find(&users)  // SELECT * FROM users

// ✅ 只查询需要的字段
var users []User
db.Select("id, name, email").Find(&users)

// ❌ N+1 查询
var users []User
db.Find(&users)
for _, user := range users {
    db.Preload("Posts").First(&user)  // N 次查询
}

// ✅ 预加载
var users []User
db.Preload("Posts").Find(&users)

// ✅ 使用 Limit 限制
db.Limit(100).Find(&users)

// ✅ 索引提示
db.Clauses(hints.UseIndex("idx_user_email")).Where("email = ?", email).Find(&user)
```

### HTTP 客户端优化

```go
// ❌ 每次请求创建新客户端
for _, url := range urls {
    resp, _ := http.Get(url)  // 无连接复用
}

// ✅ 复用客户端
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

for _, url := range urls {
    resp, _ := client.Get(url)  // 连接复用
}
```

### Profiling 与诊断

```go
import (
    _ "net/http/pprof"
    "runtime/pprof"
)

// 在代码中添加
func init() {
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
}
```

```bash
# CPU Profile
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存 Profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine Profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 生成火焰图
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile
```

### 基准测试对比

```go
// 字符串拼接对比
func BenchmarkStringConcat(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := ""
        for j := 0; j < 100; j++ {
            result += "hello"
        }
    }
}

func BenchmarkStringsBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var builder strings.Builder
        for j := 0; j < 100; j++ {
            builder.WriteString("hello")
        }
        builder.String()
    }
}

// 运行: go test -bench=BenchmarkString -benchmem
```

---

## 17. 安全实践

### Java Spring Security

```java
// Java: Spring Security 配置
@Configuration
@EnableWebSecurity
public class SecurityConfig {
    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .csrf().disable()
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/public/**").permitAll()
                .requestMatchers("/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }
}
```

### Go 安全实践

#### 1. JWT 认证

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// JWT 密钥配置
var jwtSecret = []byte("your-secret-key")

// Claims 定义
type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

// 生成 Token
func GenerateToken(userID, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// 验证 Token
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
        return jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

// Middleware
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "missing token"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // 将用户信息存入 context
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}

// 使用
func main() {
    r := gin.Default()
    r.POST("/login", loginHandler)

    // 需要认证的路由
    protected := r.Group("/api")
    protected.Use(JWTAuthMiddleware())
    {
        protected.GET("/profile", profileHandler)
    }
}
```

#### 2. 密码加密

```go
import "golang.org/x/crypto/bcrypt"

// 加密密码
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// 验证密码
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

#### 3. SQL 注入防护

```go
// ❌ 危险：字符串拼接（SQL 注入）
func BadQuery(userID string) {
    query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
    db.Exec(query)  // 可能被注入
}

// ✅ 安全：参数化查询
func GoodQuery(userID string) {
    // GORM 自动参数化
    var user User
    db.Where("id = ?", userID).First(&user)

    // 原生 SQL 也用 ?
    db.Raw("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
}

// ✅ 更安全的：白名单验证
func QueryWithWhiteList(field, value string) {
    allowedFields := map[string]bool{
        "id":    true,
        "email": true,
        "name":  true,
    }

    if !allowedFields[field] {
        return errors.New("invalid field")
    }

    db.Where(fmt.Sprintf("%s = ?", field), value).First(&User{})
}
```

#### 4. XSS 防护

```go
import "html"

// 转义 HTML 特殊字符
func SafeHTML(userInput string) string {
    return html.EscapeString(userInput)
}

// JSON 响应时，Gin 默认会处理
func SafeJSONHandler(c *gin.Context) {
    // Gin 的 c.JSON 会自动转义
    c.JSON(200, gin.H{"message": "<script>alert('xss')</script>"})
    // 输出: {"message": "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"}
}

// CSP 中间件
func CSPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")
        c.Next()
    }
}
```

#### 5. 速率限制

```go
import "github.com/ulule/limiter/v3"
import "github.com/ulule/limiter/v3/drivers/store/memory"

// 内存存储
store := memory.NewStore()

// 配置限制规则
rate := limiter.Rate{
    Period: time.Minute,
    Limit:  100, // 每分钟 100 请求
}

 limiter := limiter.New(store, rate)

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 按 IP 限流
        key := c.ClientIP()
        context, err := limiter.Get(c.Request.Context(), key)

        if err != nil {
            c.Next()
            return
        }

        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rate.Limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))

        if context.Reached {
            c.JSON(429, gin.H{"error": "too many requests"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 6. HTTPS 与 TLS

```go
// 生产环境强制 HTTPS
func ForceHTTPS() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.TLS == nil {
            // 重定向到 HTTPS
            httpsURL := "https://" + c.Request.Host + c.Request.URL.String()
            c.Redirect(301, httpsURL)
            c.Abort()
        }
        c.Next()
    }
}

// TLS 配置
certFile := "cert.pem"
keyFile := "key.pem"

r.RunTLS(":443", certFile, keyFile)

// 或使用 Let's Encrypt 自动证书
// github.com/niclabs/orb
```

#### 7. CORS 配置

```go
import "github.com/gin-contrib/cors"

func CORSConfig() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

#### 8. 安全 Headers

```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Next()
    }
}
```

### 敏感信息管理

```go
// ❌ 危险：硬编码密钥
var apiKey = "sk-1234567890abcdef"

// ✅ 安全：从环境变量读取
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY not set")
}

// ✅ 更安全：使用 Vault 或 AWS Secrets Manager
import "github.com/hashicorp/vault/api"

func GetSecret(path, key string) (string, error) {
    client, _ := vault.NewClient(vault.DefaultConfig())

    secret, err := client.KVv2("secret").Get(context.Background(), path)
    if err != nil {
        return "", err
    }

    return secret.Data[key].(string), nil
}
```

---

## 18. CI/CD 持续集成与部署

### Java (GitHub Actions + Maven)

```yaml
# .github/workflows/maven.yml
name: Java CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up JDK 17
        uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'

      - name: Cache Maven packages
        uses: actions/cache@v3
        with:
          path: ~/.m2/repository
          key: ${{ runner.os }}-m2-${{ hashFiles('**/pom.xml') }}

      - name: Build with Maven
        run: mvn clean package -DskipTests

      - name: Run tests
        run: mvn test

      - name: Build Docker image
        run: |
          docker build -t myapp:${{ github.sha }} .
          docker push myapp:${{ github.sha }}
```

### Go (GitHub Actions)

```yaml
# .github/workflows/go.yml
name: Go CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # ========== 代码检查 ==========
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  # ========== 测试 ==========
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out

  # ========== 构建 ==========
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build
        run: |
          CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: binary
          path: main

  # ========== Docker 构建 ==========
  docker:
    name: Docker Build & Push
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'

    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=sha,prefix={{branch}}-
            type=semver,pattern={{version}}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  # ========== 部署到 K8s ==========
  deploy:
    name: Deploy to Kubernetes
    runs-on: ubuntu-latest
    needs: [docker]
    if: github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v4

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
          echo "KUBECONFIG=$(pwd)/kubeconfig" >> $GITHUB_ENV

      - name: Deploy to cluster
        run: |
          kubectl set image deployment/app \
            app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} \
            --namespace=production
          kubectl rollout status deployment/app -n=production
```

### Docker 配置优化

```dockerfile
# 多阶段构建 + 优化
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 依赖分层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \      # 去除调试信息和符号表，减小体积
    -installsuffix cgo \
    -o main .

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安全：使用非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 只复制二进制
COPY --from=builder /build/main .
COPY --from=builder /build/config ./config

# 设置权限
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./main"]
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - lint
  - test
  - build
  - deploy

variables:
  GO_VERSION: "1.21"
  DOCKER_IMAGE: $CI_REGISTRY_IMAGE

.go-lint:
  image: golangci/golangci-lint:latest
  stage: lint
  script:
    - golangci-lint run ./...
  allow_failure: false

.go-test:
  image: golang:${GO_VERSION}-alpine
  stage: test
  script:
    - apk add --no-cache git
    - go mod download
    - go test -v -race -coverprofile coverage.out ./...
  coverage: '/total:.*\s(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out

docker-build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker build -t $DOCKER_IMAGE:$CI_COMMIT_SHA .
    - docker push $DOCKER_IMAGE:$CI_COMMIT_SHA
    - docker tag $DOCKER_IMAGE:$CI_COMMIT_SHA $DOCKER_IMAGE:latest
    - docker push $DOCKER_IMAGE:latest
  only:
    - main

deploy-production:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl config set-cluster k8s --server=$K8S_SERVER --token=$K8S_TOKEN
    - kubectl set image deployment/app app=$DOCKER_IMAGE:$CI_COMMIT_SHA
    - kubectl rollout status deployment/app -n=production
  environment:
    name: production
    url: https://example.com
  only:
    - main
```

### Makefile（简化操作）

```makefile
# Makefile
.PHONY: all build test lint docker run clean

GO=go
GOFLAGS=-v
BINARY_NAME=myapp
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

all: lint test build

# 构建
build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o $(BINARY_NAME)

# 测试
test:
	$(GO) test -v -race -cover ./...

# 代码检查
lint:
	golangci-lint run ./...

# Docker
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-push: docker
	docker push $(BINARY_NAME):$(VERSION)

# 运行
run: build
	./$(BINARY_NAME)

# 清理
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

# 依赖
deps:
	$(GO) mod download
	$(GO) mod tidy

# 帮助
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  docker       - Build Docker image"
	@echo "  run          - Build and run"
	@echo "  clean        - Clean build artifacts"
```

---

## 19. 日志治理与可观测性

### Java (Logback/SLF4J)

```java
// Java: Logback 配置
@Slf4j
public class UserService {
    public void createUser(User user) {
        log.info("Creating user: {}", user.getEmail());
        try {
            // 业务逻辑
            log.info("User created successfully");
        } catch (Exception e) {
            log.error("Failed to create user", e);
        }
    }
}
```

### Go 结构化日志

```go
import "log/slog"

// 使用 slog（Go 1.21+ 标准库）
func main() {
    // JSON 格式（适合收集到 ELK）
    log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    log.Info("starting server",
        "host", "localhost",
        "port", 8080,
        "env", "production",
    )
}

// 自定义 Logger
func InitLogger() *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     slog.LevelInfo,
        AddSource: true,
    }))
}

// 在业务中使用
type UserService struct {
    log *slog.Logger
    repo *UserRepo
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    s.log.InfoContext(ctx, "creating user",
        "email", user.Email,
        "request_id", getRequestID(ctx),
    )

    if err := s.repo.Create(user); err != nil {
        s.log.ErrorContext(ctx, "failed to create user",
            "error", err,
            "email", user.Email,
        )
        return err
    }

    s.log.InfoContext(ctx, "user created",
        "user_id", user.ID,
    )
    return nil
}
```

### Zap Logger（高性能）

```go
import "go.uber.org/zap"

// 创建 logger
logger, _ := zap.NewProduction()
defer logger.Sync()

// 结构化日志
logger.Info("failed to fetch URL",
    zap.String("url", "http://example.com"),
    zap.Int("attempt", 3),
    zap.Duration("backoff", time.Second),
)

// 使用 sugar logger（更灵活的 API）
sugar := logger.Sugar()
sugar.Infow("failed to fetch URL",
    "url", "http://example.com",
    "attempt", 3,
)

sugar.Infof("failed to fetch %s %d times", "URL", 3)
```

### 请求日志中间件

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

var logger *zap.Logger

func RequestLoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 生成请求 ID
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        // 记录开始时间
        start := time.Now()
        path := c.Request.URL.Path

        // 处理请求
        c.Next()

        // 记录结束
        logger.Info("request completed",
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", time.Since(start)),
            zap.String("client_ip", c.ClientIP()),
        )
    }
}

// 使用
func main() {
    r := gin.New()
    r.Use(RequestLoggerMiddleware())
}
```

### 统一响应日志

```go
// 统一日志封装
type ResponseLogger struct {
    logger *zap.Logger
}

func (rl *ResponseLogger) LogError(c *gin.Context, err error, msg string) {
    rl.logger.Error(msg,
        zap.String("request_id", c.GetString("request_id")),
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
        zap.Error(err),
    )
}

func (rl *ResponseLogger) LogWarn(c *gin.Context, msg string, fields ...zap.Field) {
    base := []zap.Field{
        zap.String("request_id", c.GetString("request_id")),
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
    }
    rl.logger.Warn(msg, append(base, fields...)...)
}
```

### ELK Stack 集成

```go
// 输出 JSON 到 stdout，Filebeat 收集
func InitELKLogger() *zap.Logger {
    config := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
        Development: false,
        Encoding:    "json",
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "@timestamp",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "message",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:    zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }

    logger, _ := config.Build()
    return logger
}
```

### Loki + Grafana 日志方案

```go
// 使用 loki 客户端
import "github.com/grafana/loki/pkg/logcli/client"

func SendToLoki() {
    client := logcli.NewClient("https://loki.example.com", "your-api-key")

    // 发送日志
    err := client.LabelValues("job")
    // ...
}
```

### 可观测性三要素

```
┌─────────────────────────────────────────────────────┐
│                   可观测性                            │
├──────────────┬──────────────┬───────────────────────┤
│   Metrics    │    Logs      │       Traces          │
│  (指标)       │   (日志)      │      (链路追踪)        │
├──────────────┼──────────────┼───────────────────────┤
│  Prometheus  │  Loki/ELK    │   Jaeger/Zipkin       │
│  Grafana     │  Grafana     │   Grafana             │
└──────────────┴──────────────┴───────────────────────┘
```

### OpenTelemetry 集成

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracing(serviceName string) func() {
    exporter, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)

    return func() {
        tp.Shutdown(context.Background())
    }
}

// 在 HTTP 请求中使用
func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)

    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
        defer span.End()

        span.SetAttributes("http.method", c.Request.Method)
        span.SetAttributes("http.url", c.Request.URL.String())

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        span.SetAttributes("http.status", c.Writer.Status())
    }
}
```

### 日志级别最佳实践

```go
// DEBUG: 开发调试，输出详细过程
log.Debug("processing item",
    "item_id", item.ID,
    "attempt", attempt,
)

// INFO: 正常运行日志
log.Info("server started",
    "port", 8080,
    "env", "production",
)

// WARN: 警告，但不影响功能
log.Warn("cache miss",
    "key", key,
    "fallback", "database",
)

// ERROR: 错误，需要关注
log.Error("database connection failed",
    "error", err,
    "host", dbHost,
)
```

---

## 20. gRPC 实战指南

### REST vs gRPC 对比

| 特性 | REST | gRPC |
|------|------|------|
| 协议 | HTTP/1.1 | HTTP/2 |
| 数据格式 | JSON | Protocol Buffers |
| 代码生成 | Swagger/OpenAPI | protoc 插件 |
| 类型安全 | 弱（运行时检查） | 强（编译时检查） |
| 流式支持 | SSE/Long Polling | 原生双向流 |
| 性能 | 较低 | 高（protobuf 体积小、解析快） |
| 适用场景 | 外部 API、Browser | 微服务内部通信 |

### 安装 protoc 工具

```bash
# macOS
brew install protobuf

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装其他语言插件（可选）
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

### 定义 Proto 文件

```protobuf
// user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/example/genproto/user/v1;userpb";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

// 用户服务
service UserService {
    // 获取用户
    rpc GetUser(GetUserRequest) returns (GetUserResponse) {
        option (google.api.http) = {
            get: "/v1/users/{id}"
        };
    }

    // 列出用户
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
        option (google.api.http) = {
            get: "/v1/users"
        };
    }

    // 创建用户
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
        option (google.api.http) = {
            post: "/v1/users"
            body: "*"
        };
    }

    // 流式响应：获取批量用户
    rpc StreamUsers(StreamUsersRequest) returns (stream User);
}

// 用户消息
message User {
    string id = 1;
    string email = 2;
    string name = 3;
    UserRole role = 4;
    google.protobuf.Timestamp created_at = 5;
}

enum UserRole {
    USER_ROLE_UNSPECIFIED = 0;
    USER_ROLE_ADMIN = 1;
    USER_ROLE_NORMAL = 2;
}

// 请求和响应消息
message GetUserRequest {
    string id = 1;
}

message GetUserResponse {
    User user = 1;
}

message ListUsersRequest {
    int32 page_size = 1;
    string page_token = 2;
}

message ListUsersResponse {
    repeated User users = 1;
    string next_page_token = 2;
}

message CreateUserRequest {
    string email = 1;
    string name = 2;
    string password = 3;
}

message CreateUserResponse {
    User user = 1;
}

message StreamUsersRequest {
    repeated string ids = 1;
}
```

### 生成代码

```bash
# 生成 Go 代码
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    user/v1/user.proto

# 生成带 grpc-gateway 的代码（用于 REST 兼容）
protoc \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    user/v1/user.proto
```

### gRPC 服务端实现

```go
// user/v1/server.go
package user

import (
    "context"
    "errors"

    "github.com/example/genproto/user/v1/userpb"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Server struct {
    userpb.UnimplementedUserServiceServer
    repo *UserRepository
}

func NewServer(repo *UserRepository) *Server {
    return &Server{repo: repo}
}

// 实现 GetUser
func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "id is required")
    }

    user, err := s.repo.FindByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
        }
        return nil, status.Error(codes.Internal, "internal error")
    }

    return &userpb.GetUserResponse{
        User: toProtoUser(user),
    }, nil
}

// 实现 ListUsers
func (s *Server) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
    if req.PageSize == 0 {
        req.PageSize = 10
    }
    if req.PageSize > 100 {
        req.PageSize = 100
    }

    users, nextToken, err := s.repo.List(ctx, int(req.PageSize), req.PageToken)
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to list users")
    }

    protoUsers := make([]*userpb.User, len(users))
    for i, user := range users {
        protoUsers[i] = toProtoUser(&user)
    }

    return &userpb.ListUsersResponse{
        Users:          protoUsers,
        NextPageToken:  nextToken,
    }, nil
}

// 实现 CreateUser
func (s *Server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
    if req.Email == "" || req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "email and name are required")
    }

    user := &User{
        Email: req.Email,
        Name:  req.Name,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, status.Error(codes.Internal, "failed to create user")
    }

    return &userpb.CreateUserResponse{
        User: toProtoUser(user),
    }, nil
}

// 流式响应
func (s *Server) StreamUsers(req *userpb.StreamUsersRequest, stream userpb.UserService_StreamUsersServer) error {
    for _, id := range req.Ids {
        user, err := s.repo.FindByID(stream.Context(), id)
        if err != nil {
            continue
        }

        if err := stream.Send(toProtoUser(user)); err != nil {
            return err
        }
    }
    return nil
}

// 辅助函数：Model -> Proto
func toProtoUser(u *User) *userpb.User {
    return &userpb.User{
        Id:        u.ID,
        Email:     u.Email,
        Name:      u.Name,
        Role:      toProtoRole(u.Role),
        CreatedAt: timestamppb.New(u.CreatedAt),
    }
}

func toProtoRole(role UserRole) userpb.UserRole {
    switch role {
    case RoleAdmin:
        return userpb.UserRole_USER_ROLE_ADMIN
    case RoleNormal:
        return userpb.UserRole_USER_ROLE_NORMAL
    default:
        return userpb.UserRole_USER_ROLE_UNSPECIFIED
    }
}
```

### gRPC 客户端

```go
// client/main.go
package main

import (
    "context"
    "log"
    "time"

    "github.com/example/genproto/user/v1/userpb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // 连接 gRPC 服务
    conn, err := grpc.Dial(
        "localhost:8080",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(loggingInterceptor),
        grpc.WithStreamInterceptor(streamLoggingInterceptor),
    )
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    client := userpb.NewUserServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    //  unary 调用
    resp, err := client.GetUser(ctx, &userpb.GetUserRequest{Id: "123"})
    if err != nil {
        log.Fatalf("GetUser failed: %v", err)
    }
    log.Printf("User: %+v", resp.User)

    // 创建用户
    createResp, err := client.CreateUser(ctx, &userpb.CreateUserRequest{
        Email:    "test@example.com",
        Name:     "Test User",
        Password: "password123",
    })
    if err != nil {
        log.Fatalf("CreateUser failed: %v", err)
    }
    log.Printf("Created User: %+v", createResp.User)

    // 流式调用
    stream, err := client.StreamUsers(ctx, &userpb.StreamUsersRequest{
        Ids: []string{"1", "2", "3"},
    })
    if err != nil {
        log.Fatalf("StreamUsers failed: %v", err)
    }

    for {
        user, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("Stream error: %v", err)
            break
        }
        log.Printf("Received user: %+v", user)
    }
}

// 拦截器示例
func loggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
    log.Printf("gRPC method: %s, request: %+v", method, req)
    err := invoker(ctx, method, req, reply, cc)
    log.Printf("gRPC response: %+v, error: %v", reply, err)
    return err
}

func streamLoggingInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
    log.Printf("gRPC stream method: %s", method)
    return streamer(ctx, desc, cc, method, opts...)
}
```

### gRPC 服务端启动

```go
func main() {
    // 创建 TCP 监听
    lis, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    // 创建 gRPC 服务器（带拦截器）
    server := grpc.NewServer(
        grpc.UnaryInterceptor(loggingInterceptor),
        grpc.StreamInterceptor(streamLoggingInterceptor),
    )

    // 注册服务
    userpb.RegisterUserServiceServer(server, NewServer(userRepo))

    log.Printf("gRPC server listening on %s", lis.Addr())
    if err := server.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

// 服务端拦截器
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Printf("gRPC method: %s", info.FullMethod)

    resp, err := handler(ctx, req)

    if err != nil {
        log.Printf("gRPC method: %s, error: %v", info.FullMethod, err)
    }

    return resp, err
}
```

### gRPC 与 HTTP 共存 (grpc-gateway)

```go
// 生成带 gateway 的代码后，可以同时提供 REST API
// 请求会通过 gateway 转换为 gRPC 调用

// main.go
import (
    "context"
    "log"
    "net/http"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "github.com/example/genproto/user/v1/userpb"
)

func grpcGateway() {
    ctx := context.Background()
    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    }

    // 注册 gateway
    err := userpb.RegisterUserServiceHandlerFromEndpoint(
        ctx,
        mux,
        "localhost:8080",  // gRPC 服务地址
        opts,
    )
    if err != nil {
        log.Fatal(err)
    }

    httpServer := &http.Server{
        Handler: mux,
        Addr:    ":8090",
    }

    log.Printf("HTTP gateway listening on %s", httpServer.Addr)
    if err := httpServer.ListenAndServe(); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

### gRPC 错误处理

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

// 服务端返回错误
func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
    }

    user, err := s.repo.FindByID(ctx, req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
        }
        return nil, status.Error(codes.Internal, "internal server error")
    }

    return &userpb.GetUserResponse{User: toProtoUser(user)}, nil
}

// 客户端处理错误
resp, err := client.GetUser(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if !ok {
        // 非 gRPC 错误
        log.Fatalf("unknown error: %v", err)
    }

    switch st.Code() {
    case codes.NotFound:
        log.Printf("User not found: %s", st.Message())
    case codes.InvalidArgument:
        log.Printf("Invalid argument: %s", st.Message())
    case codes.Internal:
        log.Printf("Internal error: %s", st.Message())
    default:
        log.Printf("Unknown gRPC error: %v", err)
    }
}
```

### gRPC 健康检查

```go
import "google.golang.org/grpc/health"
import "google.golang.org/grpc/health/grpc_health_v1"

// 实现健康检查接口
type HealthServer struct {
    health.UnimplementedHealthServer
}

func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    return &grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}

func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
    return stream.Send(&grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    })
}

// 注册健康检查服务
grpc_health_v1.RegisterHealthServer(server, &HealthServer{})
```

### gRPC 与 Docker/K8s

```yaml
# Dockerfile
FROM alpine:3.19
WORKDIR /app
COPY server /app/server
EXPOSE 8080
ENTRYPOINT ["./server"]
```

```yaml
# K8s Service (gRPC 需要指定 appProtocol)
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
    - name: grpc
      port: 8080
      targetPort: 8080
      protocol: TCP
  type: ClusterIP
```

```yaml
# K8s EndpointSlice (支持 appProtocol)
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: user-service
  labels:
    kubernetes.io/service-name: user-service
addressType: IPv4
ports:
  - name: grpc
    appProtocol: grpc
    port: 8080
    protocol: TCP
endpoints:
  - addresses:
      - "10.0.0.1"
```

---

## 21. 代码规范与最佳实践

### Go 编码规范 (Effective Go)

#### 命名规范

```go
// 包名：小写、简短、无下划线
package user          // ✅
package user_service  // ❌

// 变量名：驼峰、简洁
userName := "test"   // ✅
strUserName := "test" // ❌

// 常量名：驼峰或全大写下划线
const MaxConnection = 100     // ✅
const MAX_CONNECTION = 100    // 也可接受
const max_connection = 100   // ❌

// 接口名：名词，单方法可以 +er
type Reader interface {}
type Handler interface {}
type Logger interface {}

// 函数名：动词
func GetUser() {}      // ✅
func FetchUser() {}    // ✅
func CreateUser() {}   // ✅

// 方法名：同函数
func (u *User) Save() error {}  // ✅
func (u *User) DoSave() error {} // ❌

// 错误变量：Err 开头
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

// Getter：不需要 Get 前缀
user.Name()    // ✅
user.GetName()  // ❌
```

#### 格式规范

```go
// 导入分组：标准库 → 第三方 → 项目内部
import (
    "context"
    "errors"
    "fmt"

    "github.com/pkg/errors"
    "go.uber.org/zap"

    "github.com/example/project/internal/model"
)

// 行长度：没有硬限制，但保持合理
// 如果一行太长，考虑重构
func longFunction() {
    // ✅ 可以接受的行长度
    result := someFunctionWithLongName(param1, param2, param3)

    // ❌ 过长，考虑分行
    result := someFunctionWithLongName(param1, param2,
        param3, param4)
}

// 注释：注释应该解释"为什么"，而不是"是什么"
func CreateUser(name string) error {
    // 使用唯一索引避免重复邮箱，而不是依赖应用层检查
    // 因为在高并发场景下，应用层检查存在竞态条件
    return db.Create(&User{Name: name})
}
```

#### 错误处理规范

```go
// ❌ 忽略错误
data, _ := os.ReadFile("config.json")

// ✅ 明确处理
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}

// ❌ 错误信息不明确
if err != nil {
    return err
}

// ✅ 错误信息包含上下文
if err != nil {
    return fmt.Errorf("create user: %w", err)
}

// ✅ 哨兵错误用于可预测的错误
if errors.Is(err, sql.ErrNoRows) {
    return ErrNotFound
}

// ✅ 自定义错误类型用于复杂场景
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

#### 并发规范

```go
// ❌ 共享变量导致竞态
var counter int
go func() { counter++ }()
go func() { counter++ }()

// ✅ 使用 Channel 通信
ch := make(chan int)
go func() { ch <- 1 }()
go func() { ch <- 2 }()

// ✅ 使用 sync.Mutex 保护共享状态
var mu sync.Mutex
var counter int

mu.Lock()
counter++
mu.Unlock()

// ✅ 使用 sync.WaitGroup 等待完成
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        doWork()
    }()
}
wg.Wait()

// ✅ Context 用于取消
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // 工作
        }
    }
}(ctx)
```

### 项目结构规范

```
├── cmd/                    # 可执行程序入口
│   └── server/
│       └── main.go
│
├── internal/               # 私有代码（仅本项目使用）
│   ├── handler/           # HTTP/RPC 处理器
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问
│   ├── model/             # 数据模型
│   └── middleware/        # 中间件
│
├── pkg/                   # 公共库（可被外部项目导入）
│   ├── validator/         # 验证器
│   └── response/          # 响应封装
│
├── api/                   # API 定义（proto 文件）
│   └── user/v1/
│       └── user.proto
│
├── configs/               # 配置文件
│   └── config.yaml
│
├── scripts/               # 脚本（数据库迁移、代码生成等）
│   └── generate.sh
│
├── test/                  # 集成测试
│   └── integration/
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

### API 设计规范 (REST)

```go
// ✅ RESTful 路由设计
// 资源用名词，动作用 HTTP 方法
// GET    /users          - 列出用户
// GET    /users/:id      - 获取单个用户
// POST   /users          - 创建用户
// PUT    /users/:id      - 更新用户（完整）
// PATCH  /users/:id      - 更新用户（部分）
// DELETE /users/:id      - 删除用户

// ❌ 避免动词在路径中
// POST /createUser        ❌
// POST /users/create       ✅

// ✅ 版本控制
// /v1/users
// /v2/users

// ✅ 使用复数名词
// GET /users              ✅
// GET /user               ❌

// ✅ 嵌套资源（适度）
// GET /users/:id/posts    ✅ 获取用户的所有文章
// 避免过深嵌套（超过2层）
```

### 配置管理规范

```go
// ❌ 硬编码配置
port := 8080
timeout := 30

// ✅ 使用 Viper 管理配置
import "github.com/spf13/viper"

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Host string
    Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
    Host     string
    Port     int    `mapstructure:"port"`
    User     string
    Password string `mapstructure:"password"`
    Name     string
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")

    // 支持环境变量覆盖
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

```yaml
# configs/config.yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 3306
  user: root
  password: ${DB_PASSWORD}  # 环境变量
  name: myapp

redis:
  host: localhost
  port: 6379
```

### 测试规范

```go
// 文件命名
user_test.go           // 单元测试
user_integration_test.go // 集成测试
user_benchmark_test.go  // 基准测试

// 测试函数命名
func TestUserService_GetUser(t *testing.T) {}       // ✅
func TestGetUser(t *testing.T) {}                  // ✅
func TestGet(t *testing.T) {}                      // ❌ 太模糊

// 测试分组
func TestUserService(t *testing.T) {
    t.Run("GetUser", func(t *testing.T) { ... })
    t.Run("CreateUser", func(t *testing.T) { ... })
    t.Run("DeleteUser", func(t *testing.T) { ... })
}

// Table-Driven Tests
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### 代码审查清单

```
□ 代码能够编译通过
□ 所有测试通过
□ 没有未处理的错误
□ 没有 fmt.Printf 用于调试
□ 变量命名清晰、符合规范
□ 注释解释"为什么"而非"是什么"
□ 错误信息包含上下文
□ 敏感信息不硬编码
□ 没有调试代码残留
□ 日志级别使用正确
□ 并发使用安全（无竞态）
□ 资源正确释放（defer, context）
□ 安全检查（输入验证、SQL注入等）
□ 性能考虑（连接池、批处理等）
□ API 响应时间合理
□ 代码复杂度适中（圈复杂度）
```

---

## 总结

| 维度 | Java | Go |
|------|------|-----|
| 类型系统 | 继承 + 泛型 | 组合 + 泛型 (1.18+) |
| 错误处理 | 异常 | error 返回值 |
| 并发 | Thread / Executor | Goroutine + Channel |
| 接口 | 显式实现 | 隐式实现 |
| 分层 | 强制的多层 | 灵活的扁平 |
| DTO | 必需 | 可选 |
| DI | 容器自动 | 构造函数手动 |
| 性能 | JIT 优化 | 编译优化 |
| 测试 | JUnit + Mockito | testing + testify |
| 缓存 | Spring Cache | 手动 + go-redis |
| 微服务 | Spring Cloud | 灵活组合生态 |
| 部署 | WAR/JAR | 二进制 / 容器 |
| 安全 | Spring Security | 中间件组合 |
| CI/CD | Jenkins / GitHub Actions | GitHub Actions / GitLab CI |
| 日志 | Logback + ELK | slog / zap + Loki |
| 内存占用 | 512MB - 2GB | 64MB - 256MB |
| 启动时间 | 10-30 秒 | < 1 秒 |

### 核心转变

1. **继承 → 组合**：用嵌入代替 extends
2. **异常 → error**：用返回值代替 try-catch
3. **Thread → Goroutine**：用轻量协程代替重量线程
4. **强制分层 → 按需分层**：Go 更灵活
5. **显式接口 → 隐式接口**：实现即满足
6. **自动注入 → 构造函数注入**：显式优于隐式
7. **注解缓存 → 手动缓存**：需要更多代码但更可控
8. **WAR 包 → 二进制/容器**：Go 部署更简单
9. **Spring Security → 中间件组合**：按需拼装
10. **Logback → slog/zap**：结构化日志
11. **REST → gRPC**：高性能微服务通信
12. **弱类型 → Protocol Buffers**：编译时类型安全

---

## 参考资源

### 官方文档
- [Go 语言官方文档](https://go.dev/doc/)
- [Go 标准库](https://pkg.go.dev/)
- [GORM 文档](https://gorm.io/docs/)
- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/zh-cn/docs/)
- [OpenTelemetry](https://opentelemetry.io/)

### 进阶阅读
- [Go 高级编程](https://chai2010.cn/advanced-go-programming-book/)
- [实效 Go 编程](https://go-proverbs.github.io/)
- [Uber Go 编码规范](https://github.com/uber-go/guide)
- [Go 项目结构推荐](https://github.com/golang-standards/project-layout)
- [Go 语言性能优化实战](https://github.com/dgryski/go-perfbook)
- [Go 安全编程指南](https://github.com/guardrailsio/awesome-golang-security)
- [gRPC 官方 Go 示例](https://github.com/grpc/grpc-go/tree/master/examples)

### 工具库
| 用途 | 推荐库 |
|------|--------|
| Web 框架 | Gin, Echo, Chi |
| ORM | GORM, sqlx |
| 验证 | go-playground/validator |
| 日志 | slog (标准库), zap, logrus |
| Redis | go-redis/redis |
| 微服务 | gRPC, DTM, consul |
| 配置 | viper |
| 测试 | testify, gomock |
| Docker | docker, compose |
| K8s | client-go |
| 安全 | golang-jwt, bcrypt |
| CI/CD | actions, gitlab-runner |
| 链路追踪 | OpenTelemetry, jaeger |

---

*文档版本: 5.0 | 更新日期: 2026-04-25*

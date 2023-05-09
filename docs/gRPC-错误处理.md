# 错误处理

## gRPC code

类似于HTTP定义了一套响应状态码，gRPC也定义有一些状态码。Go语言中此状态码由codes定义，本质上是一个uint32。

`type Code uint32`

使用时需导入`google.golang.org/grpc/codes`包

<p>目前已经定义的状态码有如下几种。</p>

<table>
<thead>
<tr>
<th>Code</th>
<th>值</th>
<th>含义</th>
</tr>
</thead>

<tbody>
<tr>
<td>OK</td>
<td>0</td>
<td>请求成功</td>
</tr>

<tr>
<td>Canceled</td>
<td>1</td>
<td>操作已取消</td>
</tr>

<tr>
<td>Unknown</td>
<td>2</td>
<td>未知错误。如果从另一个地址空间接收到的状态值属 于在该地址空间中未知的错误空间，则可以返回此错误的示例。 没有返回足够的错误信息的API引发的错误也可能会转换为此错误</td>
</tr>

<tr>
<td>InvalidArgument</td>
<td>3</td>
<td>表示客户端指定的参数无效。 请注意，这与 FailedPrecondition 不同。 它表示无论系统状态如何都有问题的参数（例如，格式错误的文件名）。</td>
</tr>

<tr>
<td>DeadlineExceeded</td>
<td>4</td>
<td>表示操作在完成之前已过期。对于改变系统状态的操作，即使操作成功完成，也可能会返回此错误。 例如，来自服务器的成功响应可能已延迟足够长的时间以使截止日期到期。</td>
</tr>

<tr>
<td>NotFound</td>
<td>5</td>
<td>表示未找到某些请求的实体（例如，文件或目录）。</td>
</tr>

<tr>
<td>AlreadyExists</td>
<td>6</td>
<td>创建实体的尝试失败，因为实体已经存在。</td>
</tr>

<tr>
<td>PermissionDenied</td>
<td>7</td>
<td>表示调用者没有权限执行指定的操作。 它不能用于拒绝由耗尽某些资源引起的（使用 ResourceExhausted ）。 如果无法识别调用者，也不能使用它（使用 Unauthenticated ）。</td>
</tr>

<tr>
<td>ResourceExhausted</td>
<td>8</td>
<td>表示某些资源已耗尽，可能是每个用户的配额，或者整个文件系统空间不足</td>
</tr>

<tr>
<td>FailedPrecondition</td>
<td>9</td>
<td>指示操作被拒绝，因为系统未处于操作执行所需的状态。 例如，要删除的目录可能是非空的，rmdir 操作应用于非目录等。</td>
</tr>

<tr>
<td>Aborted</td>
<td>10</td>
<td>表示操作被中止，通常是由于并发问题，如排序器检查失败、事务中止等。</td>
</tr>

<tr>
<td>OutOfRange</td>
<td>11</td>
<td>表示尝试超出有效范围的操作。</td>
</tr>

<tr>
<td>Unimplemented</td>
<td>12</td>
<td>表示此服务中未实施或不支持/启用操作。</td>
</tr>

<tr>
<td>Internal</td>
<td>13</td>
<td>意味着底层系统预期的一些不变量已被破坏。 如果你看到这个错误，则说明问题很严重。</td>
</tr>

<tr>
<td>Unavailable</td>
<td>14</td>
<td>表示服务当前不可用。这很可能是暂时的情况，可以通过回退重试来纠正。 请注意，重试非幂等操作并不总是安全的。</td>
</tr>

<tr>
<td>DataLoss</td>
<td>15</td>
<td>表示不可恢复的数据丢失或损坏</td>
</tr>

<tr>
<td>Unauthenticated</td>
<td>16</td>
<td>表示请求没有用于操作的有效身份验证凭据</td>
</tr>

<tr>
<td>_maxCode</td>
<td>17</td>
<td>-</td>
</tr>
</tbody>
</table>

## gRPC Status

gRPC Status 定义在`google.golang.org/grpc/status`，使用时需导入

RPC服务的方法应该返回 `nil` 或来自`status.Status`类型的错误。客户端可以直接访问错误。

### 创建错误

当遇到错误时，gRPC服务的方法函数应该创建一个 `status.Status`。通常我们会使用 status.New函数并传入适当的`status.Code`和错误描述来生成一个`status.Status`。调用`status.Err`方法便能将一个`status.Status`转为`error`类型。也存在一个简单的`status.Error`方法直接生成`error`。下面是两种方式的比较。

```go
    // 创建status.Status
    st := status.New(codes.NotFound, "some description")
    err := st.Err()  // 转为error类型
    
    // vs.
    
    err := status.Error(codes.NotFound, "some description")
```

### 为错误添加其他详细信息

在某些情况下，可能需要为服务器端的特定错误添加详细信息。`status.WithDetails`就是为此而存在的，它可以添加任意多个proto.Message，我们可以使用`google.golang.org/genproto/googleapis/rpc/errdetails`中的定义或自定义的错误详情。

```go
    st := status.New(codes.ResourceExhausted, "Request limit exceeded.")
    ds, _ := st.WithDetails(
        // proto.Message
    )
    return nil, ds.Err()

```
然后，客户端可以通过首先将普通`error`类型转换回`status.Status`，然后使用`status.Details`来读取这些详细信息。

```go
    s := status.Convert(err)
    for _, d := range s.Details() {
        // ...
    }
```


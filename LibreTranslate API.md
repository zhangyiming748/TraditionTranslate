---
title: 默认模块
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# 默认模块

Base URLs:

# Authentication

# Default

## POST 检测文本语言

POST /detect

> Body 请求参数

```yaml
q: What language is this?
api_key: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» q|body|string| 是 |none|
|» api_key|body|string| 否 |none|

> 返回示例

> 200 Response

```json
[
    {
        "confidence": 100.0,
        "language": "en"
    }
]
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» confidence|integer|false|none||none|
|» language|string|false|none||none|

## GET 获取受支持的语言

GET /languages

> 返回示例

> 200 Response

```
[{"code":"en","name":"English","targets":["de","en","ja","zh-Hans"]},{"code":"zh-Hans","name":"Chinese","targets":["de","en","ja","zh-Hans"]},{"code":"de","name":"German","targets":["de","en","ja","zh-Hans"]},{"code":"ja","name":"Japanese","targets":["de","en","ja","zh-Hans"]}]
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## POST 翻译文本

POST /translate

> Body 请求参数

```yaml
q: ""
source: auto
target: zh
format: text
alternatives: 3
api_key: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» q|body|string| 是 |待翻译文本|
|» source|body|string| 是 |源语言代码或选择“自动”来自动检测|
|» target|body|string| 是 |目标语言代码|
|» format|body|string| 否 |源文本格式|
|» alternatives|body|integer| 否 |首选替代翻译数量|
|» api_key|body|string| 否 |API 密钥|

#### 详细说明

**» format**: 源文本格式
text - 纯文本
html - HTML 标记

> 返回示例

> 200 Response

```
{"detectedLanguage":{"confidence":100,"language":"en"},"translatedText":"大胸和小猫咪"}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## POST 翻译一个文件

POST /translate_file

> Body 请求参数

```yaml
file: file:///Users/zen/Documents/post.txt
source: auto
target: zh
api_key: ""

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» file|body|string(binary)| 是 |none|
|» source|body|string| 否 |none|
|» target|body|string| 否 |none|
|» api_key|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "translatedFileUrl": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» translatedFileUrl|string|true|none||none|

## GET 获取前端设置

GET /frontend/settings

> 返回示例

> 200 Response

```json
{
    "apiKeys": false,
    "charLimit": -1,
    "filesTranslation": true,
    "frontendTimeout": 500,
    "keyRequired": false,
    "language": {
        "source": {
            "code": "auto",
            "name": "Auto Detect"
        },
        "target": {
            "code": "zh",
            "name": "Chinese"
        }
    },
    "suggestions": false,
    "supportedFilesFormat": [
        ".txt",
        ".odt",
        ".odp",
        ".docx",
        ".pptx",
        ".epub",
        ".html",
        ".srt",
        ".pdf"
    ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» apiKeys|boolean|true|none||none|
|» charLimit|integer|true|none||none|
|» filesTranslation|boolean|true|none||none|
|» frontendTimeout|integer|true|none||none|
|» keyRequired|boolean|true|none||none|
|» language|object|true|none||none|
|»» source|object|true|none||none|
|»»» code|string|true|none||none|
|»»» name|string|true|none||none|
|»» target|object|true|none||none|
|»»» code|string|true|none||none|
|»»» name|string|true|none||none|
|» suggestions|boolean|true|none||none|
|» supportedFilesFormat|[string]|true|none||none|

## GET 健康监测

GET /health

> 返回示例

> 200 Response

```json
{
    "status": "ok"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» status|string|true|none||none|

## POST 提交一份建议以改进翻译

POST /suggest

> Body 请求参数

```yaml
q: hello
s: 你好
source: en
target: zh

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 是 |none|
|» q|body|string| 是 |原文|
|» s|body|string| 是 |建议译文|
|» source|body|string| 是 |原文语言|
|» target|body|string| 是 |建议译文语言|

> 返回示例

> 200 Response

```
{"error":"Suggestions are disabled on this server."}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 数据模型


# Базовый сервис на GrathQL

Небольшой GraphQL-сервис для управления сущностью **Main** и её вложенными коллекциями **Tools**, **Tables**, **Chairs**.

## Особенности
- Первичный ключ `main.id` — UUID (генерируется автоматически, либо задаётся вручную).
- Дочерние таблицы `tools`, `tables`, `chairs` автоматически синхронизируются через JSON-поле `sub_obj` и триггеры.
- Soft-delete через поле `deleted_at`.

## Запуск
```bash
docker compose -up --build
```
По умолчанию сервер слушает `http://localhost:8080/`.

## GraphQL Playground
Откройте в браузере:
```
http://localhost:8080/
```
Интерактивная консоль для отправки запросов.

## Доступные операции

### 1. Создание записи
```graphql
mutation {
  createMain(input: {
    id: "550e8400-e29b-41d4-a716-446655440000"  # либо ""
    title: "My Project"
    subId: 123
    tools: [{ title: "Hammer", description: "Steel hammer" }]
    tables: [{ name: "Workbench" }]
    chairs: [{ name: "Stool", type: "abc" }]
  }) {
    id
    title
    tools { id title description }
    tables { id name }
    chairs { id name type }
    createdAt
    updatedAt
  }
}
```

### 2. Получение по ID
```graphql
{
  getMain(id: "550e8400-e29b-41d4-a716-446655440000") {
    id
    title
    subId
    tools { id title description }
    tables { id name }
    chairs { id name type }
    createdAt
    updatedAt
    deletedAt
  }
}
```

### 3. Обновление
```graphql
mutation {
  updateMain(
    id: "550e8400-e29b-41d4-a716-446655440000",
    input: {
      title: "My Project (v2)"
      subId: 456
      tools: []
      tables: [{ name: "Conference Table" }]
      chairs: []
    }
  ) {
    id
    title
    tables { id name }
    updatedAt
  }
}
```

### 4. Удаление (soft-delete)
```graphql
mutation {
  deleteMain(id: "550e8400-e29b-41d4-a716-446655440000") {
    deleteId
  }
}
```
- Если нет что удалять, вернётся ошибка `nothing to delete for id …`.

### 5. Удаление по отдельности (пример)
```graphql
mutation RemoveAllTools {
  updateMain(
    id: "ec50f4c8-b776-4747-aab9-fe45d8c8fc2d",
    input: {
      id:    "ec50f4c8-b776-4747-aab9-fe45d8c8fc2d"
      title: "Текущий заголовок"   # замените на реальный
      subId: null                  # или текущее значение
      tools: []                    # <-- удаляем все tools
      tables: [                    # <-- нужно передать существующие tables
        { name: "Workbench" }
      ]
      chairs: [                    # <-- и существующие chairs
        { name: "Stool", type: "abc" }
      ]
    }
  ) {
    id
    tools {
      id
      title
    }
  }
}
```

## Модель данных
```graphql
input MainInput {
  id: ID!           # UUID
  title: String!
  subId: Int
  tools: [ToolInput!]
  tables: [TableInput!]
  chairs: [ChairInput!]
}

type Main {
  id: ID!
  title: String!
  subId: Int
  tools: [Tool!]
  tables: [Table!]
  chairs: [Chair!]
  createdAt: String!
  updatedAt: String!
  deletedAt: String
}
```

## Примечания
- `chair.type`: ENUM со значениями `"abc"` или `"cde"`.
- Временные метки в формате ISO-8601.
- При пустых массивах дочерние записи очищаются.
- Заранее сформирован env файл как пример.


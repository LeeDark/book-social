Page navigation contract

- backend routes
- handlers
- templates, forms, layouts
- navigation, links
- redirects, htmx/partial rendering

BookSocial

```text
GET /static/*       -> Static files
GET /               -> Home page
GET /about          -> About page
GET /books          -> Catalog/List books page
GET /books/{slug}   -> Book details page
GET /authors/{id}   -> Author page
```

```text
GET /   -> Home page
- route: GET /
- active nav item: "home"
- template: home/index.html
```

```text
GET /books
- full page by default
- return catalog page with layout
- active nav item: "catalog"
- page title: "Books"
- supports query params:
  - ?q=
  - ?page=
  - ?genre=
- if HR-Request: true, returns only books list fragment
- response must include Vary: HX-Request
```


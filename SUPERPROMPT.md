# Pollex — Contexto completo para nueva sesión

> **Copia todo este archivo y pégalo como primer mensaje en la nueva sesión de Claude Code abierta en `~/Projects/pollex/`.**
> Después de usarlo, puedes borrar este archivo.

---

## Qué es Pollex

Herramienta para "pulir" texto en inglés — corrige gramática, sintaxis y coherencia. El resultado debe sonar como un hablante fluido de inglés como segunda lengua: profesional y claro, pero NO como IA ni como nativo perfecto.

## Arquitectura (3 capas, ya decidida)

```
┌──────────────────────────────────────────────────┐
│                   MI MÁQUINA                      │
│  ┌─────────────────────┐                         │
│  │  Browser Extension  │  Manifest V3            │
│  │  (Popup UI)         │  Chrome/Brave/Firefox   │
│  └─────────┬───────────┘                         │
│            │ HTTP (JSON)                         │
└────────────┼─────────────────────────────────────┘
             │  LAN (http://jetson.local:8090)
┌────────────┼─────────────────────────────────────┐
│            ▼          JETSON NANO 4GB             │
│  ┌─────────────────────┐                         │
│  │  Pollex API (Go)    │  :8090                  │
│  │  - /api/polish      │  Binario estático ~10MB │
│  │  - /api/models      │  RAM: ~15MB             │
│  │  - /api/health      │                         │
│  └────┬──────────┬─────┘                         │
│       ▼          ▼                               │
│  ┌─────────┐  ┌────────────────┐                 │
│  │ Ollama  │  │ Claude API     │                 │
│  │ :11434  │  │ (comparación)  │                 │
│  └─────────┘  └────────────────┘                 │
└──────────────────────────────────────────────────┘
```

**Capa 1 — Extension:** Solo UI. Sin API keys ni lógica de negocio. Envía texto al backend, muestra resultado.

**Capa 2 — Go API:** El cerebro. Recibe texto + modelo, aplica system prompt, rutea a Ollama o Claude, devuelve respuesta normalizada. Un solo binario estático.

**Capa 3 — LLM backends:** Ollama (local en Jetson) + Claude API (externo, solo para comparar calidad).

## Hardware: Jetson Nano 4GB

Restricción crítica. 4GB compartidos CPU/GPU:

| Componente | RAM |
|---|---|
| JetPack OS headless | ~500MB |
| Ollama runtime | ~200MB |
| Modelo activo (Qwen2.5-1.5B Q4) | ~1.0GB |
| Pollex Go API | ~15MB |
| **Total** | **~1.7GB** |
| **Libre** | **~2.3GB** |

Modelo principal: **Qwen2.5-1.5B** (mejor equilibrio calidad/recurso). CUDA 10.2, ARM64, 128 CUDA cores Maxwell.

## Go API — Diseño del Backend

Todo en `package main`, sin subdirectorios `internal/`/`pkg/`. ~700-800 líneas total.

**Patrón central — Adapter interface:**
```go
type LLMAdapter interface {
    Name() string
    Polish(ctx context.Context, text string, systemPrompt string) (string, error)
    Available() bool
}
```

Cada backend (Ollama, Claude) implementa esta interfaz. `main.go` tiene un `map[string]LLMAdapter`. Añadir un nuevo backend = implementar la interfaz + registrarlo.

**Endpoints:**

| Método | Ruta | Request | Response |
|---|---|---|---|
| POST | /api/polish | `{"text":"...","model_id":"qwen2.5:1.5b"}` | `{"polished":"...","model":"...","elapsed_ms":12340}` |
| GET | /api/models | — | `[{"id":"qwen2.5:1.5b","name":"Qwen 2.5 1.5B","provider":"ollama"}]` |
| GET | /api/health | — | `{"status":"ok"}` |

**Timeouts escalonados** (la Jetson tarda 10-30s por respuesta):
- Extension AbortController: 70s
- Go http.TimeoutHandler: 65s
- Adapter HTTP client: 60s

**CORS:** `Access-Control-Allow-Origin: *` (uso en LAN).

**Estructura backend:**
```
backend/
├── go.mod
├── main.go              # Entry point, wiring, flag --mock
├── config.go            # Config desde YAML + env vars
├── config.yaml.example
├── adapter.go           # Interface LLMAdapter
├── adapter_mock.go      # Respuestas simuladas con delay configurable
├── adapter_ollama.go    # Ollama (OpenAI-compatible API)
├── adapter_claude.go    # Claude (Anthropic API)
├── handler_polish.go    # POST /api/polish
├── handler_models.go    # GET /api/models
├── handler_health.go    # GET /api/health
├── middleware.go        # CORS + logging + timeout
└── *_test.go            # Table-driven, httptest
```

## Browser Extension — Diseño del Frontend

Manifest V3. Permisos mínimos: solo `storage`.

```
extension/
├── manifest.json
├── popup.html / .css / .js
├── api.js               # Cliente HTTP compartido
├── settings.html / .js  # URL del API + Test Connection
└── icons/               # 16/48/128px
```

Flujo: Click icono → popup → pegar texto → seleccionar modelo → "Polish" → spinner con timer en vivo ("Polishing... 23s") → resultado + botón Copy.

Settings: campo URL del API (default `http://localhost:8090`), botón "Test Connection" → `/api/health`. Persiste en `chrome.storage.local`.

## System Prompt

Ya existe en `prompts/polish.txt`. 9 reglas. Si los modelos de 1.5B no siguen bien el prompt largo, crear versión reducida (4-5 reglas).

## Estado actual del proyecto

El scaffolding ya está creado:
```
pollex/
├── .gitignore
├── Makefile              # Targets listos (dev, test, build, deploy...)
├── tasks/
│   ├── todo.md           # Checkboxes de las 5 fases
│   └── lessons.md        # Vacío
├── backend/
│   └── go.mod            # module pollex/backend, go 1.22
├── extension/
│   └── .gitkeep
├── prompts/
│   └── polish.txt        # System prompt listo
└── deploy/
    └── .gitkeep
```

**Git NO está inicializado todavía.** Necesitas hacer `git init` + commit inicial antes de empezar.

## Tarea inmediata: Paso 0 + Fase 1

### Paso 0 — Setup inicial (hacer PRIMERO)

1. **Inicializar git:**
   ```sh
   git init
   git add .
   git commit -m "chore: project scaffolding"
   ```

2. **Crear documentación en el Obsidian vault** (NO en el repo).
   El vault está en `~/Projects/knowledge/`. La documentación de Pollex va en `~/Projects/knowledge/10_projects/pollex/`. **No existe nada todavía, hay que crearlo desde cero.**

   Crear esta estructura:
   ```
   ~/Projects/knowledge/10_projects/pollex/
   ├── _index.md                              # Ficha del proyecto
   ├── adrs/
   │   └── 001-local-llm-on-jetson-nano.md    # ADR: por qué modelo local en Jetson
   ├── runbooks/
   │   ├── deploy-jetson.md                    # Cómo desplegar paso a paso
   │   └── flash-jetson.md                     # Cómo flashear JetPack desde cero
   └── troubleshooting/
       └── jetson-memory.md                    # Problemas de memoria y soluciones
   ```

   **`_index.md`** debe incluir: nombre del proyecto, descripción corta, stack (Go 1.22, Manifest V3, Ollama), hardware target (Jetson Nano 4GB), estructura del repo, status actual (en desarrollo, Fase 1).

   **`001-local-llm-on-jetson-nano.md`** (ADR): decisión de usar modelos locales en Jetson Nano en vez de solo APIs cloud. Contexto: privacidad, latencia en LAN, coste cero por inferencia. Consecuencias: limitado a modelos 1-3B, respuestas más lentas (10-30s), calidad inferior a Claude.

   **Los runbooks y troubleshooting** pueden tener contenido placeholder por ahora — se llenarán cuando hagamos la Fase 5 (deploy).

### Fase 1 — Backend Core (la implementación real)

Seguir TDD estricto: test primero (red), implementar después (green). Go standards: table-driven tests con `t.Run`, `context.Context` en todo I/O, error wrapping `fmt.Errorf("componente: %w", err)`, `go vet` limpio.

Orden:
1. `config.go` + `config_test.go` — Carga config YAML + env vars (puerto, Ollama URL, Claude API key, ruta del prompt)
2. `adapter.go` — Interface `LLMAdapter`
3. `adapter_mock.go` — Respuestas simuladas con delay configurable (para desarrollo sin Ollama)
4. `adapter_ollama.go` + `adapter_ollama_test.go` — Cliente HTTP a Ollama (OpenAI-compatible API en `:11434`). Tests con `httptest`
5. `handler_health.go` + `handler_models.go` + `handler_polish.go` + tests — Handlers HTTP. Polish recibe `{"text":"...","model_id":"..."}`, busca adapter en el map, llama a `Polish()`, devuelve resultado
6. `middleware.go` — CORS + request logging + timeout
7. `main.go` — Wiring: carga config, crea adapters, registra rutas, arranca servidor. Flag `--mock` para usar mock adapter
8. **Verificar:** `make test` pasa, `curl localhost:8090/api/health` funciona, `curl -X POST localhost:8090/api/polish -d '{"text":"i goes to store","model_id":"mock"}'` devuelve texto pulido

Marcar items en `tasks/todo.md` conforme se completan. Si hay correcciones o aprendizajes, registrarlos en `tasks/lessons.md`.

## Reglas importantes

- **Documentación operacional (ADRs, runbooks, troubleshooting)** → vault (`~/Projects/knowledge/10_projects/pollex/`), NUNCA en el repo
- **Tasks y code tracking** → repo (`tasks/todo.md`, `tasks/lessons.md`)
- **Go:** stdlib `net/http`, NO frameworks. `package main`, NO subdirectorios
- **No over-engineering:** El backend entero son ~700-800 líneas. Sin `internal/`, sin `pkg/`, sin interfaces innecesarias
- **Makefile ya existe** con todos los targets definidos
- **El system prompt ya existe** en `prompts/polish.txt`

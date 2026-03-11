# 📚 Documentación EduGo API Mobile

¡Bienvenido al núcleo de la experiencia de usuario! La **EduGo API Mobile** es la máquina que impulsa las aplicaciones móviles y web de EduGo, proveyendo acceso ultra-rápido y seguro a materiales educativos, sistemas de evaluaciones dinámicas, métricas de progreso y pantallas UI administradas desde la nube.

> [!TIP]
> Antes de adentrarte en el código, te recomendamos encarecidamente revisar la **[Arquitectura](architecture.md)** para entender cómo están distribuidas las responsabilidades.

---

## 🧭 Índice de Navegación

### 1. 🚁 Visión General y Setup
* 📐 **[Arquitectura y Diseño (Clean Architecture)](architecture.md)** — Reglas de diseño, capas internas y diagramas de flujo.
* ⚙️ **[Detalles y Stack Técnico](technical.md)** — Lenguaje, librerías, middlewares y estrategias de testing.
* 🌍 **[Integración con el Ecosistema](ecosystem.md)** — Cómo encaja esta API en la galaxia de servicios de EduGo.

---

### 2. 🧠 Entidades de Negocio (El Corazón del Sistema)
Nuestros flujos de negocio están desacoplados en dominios claros. Explora cada uno para entender sus Modelos de Datos, Endpoints y DTOs asociados:

* 📚 **[Materiales (Material)](entities/material.md)** — Subida, bajada y gestión de recursos de estudio (Videos, PDFs, interactivos).
* 📝 **[Evaluaciones (Assessment)](entities/assessment.md)** — Del diseño del examen hasta el algoritmo de calificación progresiva.
* 📈 **[Progreso (Progress)](entities/progress.md)** — Motores de seguimiento de avance del estudiante.
* 🎨 **[Pantallas Dinámicas (Screen)](entities/screen.md)** — Inyección de UI en tiempo real para aplicaciones móviles (Server-Driven UI).
* 📊 **[Estadísticas (Stats)](entities/stats.md)** — Agregaciones de uso y rendimiento consumidas por los tutores.
* 👨‍👩‍👧 **[Tutores (Guardian)](entities/guardian.md)** — Vínculos de supervisión entre los acudientes y los estudiantes.

---
*Mantenido con ❤️ por el equipo de ingeniería de EduGo.*

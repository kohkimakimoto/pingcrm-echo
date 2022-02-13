import { createApp, h } from 'vue'
import { InertiaProgress } from '@inertiajs/progress'
import { createInertiaApp } from '@inertiajs/inertia-vue3'
import '../css/app.css'

InertiaProgress.init()

createInertiaApp({
  resolve: async name => {
    if (import.meta.env.DEV) {
      return await import(`./Pages/${name}.vue`)
    } else {
      let pages = import.meta.glob('./Pages/**/*.vue')
      const importPage = pages[`./Pages/${name}.vue`]
      return importPage().then(module => module.default)
    }
  },
  title: title => title ? `${title} - Ping CRM` : 'Ping CRM',
  setup({ el, App, props, plugin }) {
    createApp({ render: () => h(App, props) })
      .use(plugin)
      .mount(el)
  },
})

type Theme = 'light' | 'dark'

export const useTheme = () => {
  const theme = useState<Theme>('theme', () => 'dark')

  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme

    if (process.client) {
      const html = document.documentElement
      html.classList.remove('light', 'dark')
      html.classList.add(newTheme)
      localStorage.setItem('hub-theme', newTheme)
    }
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  const initTheme = () => {
    if (!process.client) return

    const savedTheme = localStorage.getItem('hub-theme') as Theme | null
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

    const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light')
    setTheme(initialTheme)
  }

  return {
    theme: readonly(theme),
    setTheme,
    toggleTheme,
    initTheme,
  }
}

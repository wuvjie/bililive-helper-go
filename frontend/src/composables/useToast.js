let toastContainer = null

function getContainer() {
  if (!toastContainer) {
    toastContainer = document.createElement('div')
    toastContainer.id = 'toasts'
    toastContainer.style.cssText = 'position:fixed;top:24px;right:24px;z-index:300;display:flex;flex-direction:column;gap:12px;pointer-events:none;'
    document.body.appendChild(toastContainer)
  }
  return toastContainer
}

export function useToast() {
  function toast(message, type = 'info') {
    const container = getContainer()
    const el = document.createElement('div')
    el.className = `toast toast-${type}`
    el.textContent = message
    el.style.cssText = 'padding:12px 20px;border-radius:8px;font:500 13px var(--font-sans);color:#fff;pointer-events:auto;animation:tin 0.3s cubic-bezier(0.16, 1, 0.3, 1),tout 0.3s ease 3s forwards;box-shadow:0 8px 16px rgba(0,0,0,0.1);'

    if (type === 'ok') {
      el.style.background = 'var(--text)'
    } else if (type === 'err') {
      el.style.background = 'var(--err)'
    } else {
      el.style.background = 'var(--info)'
    }

    container.appendChild(el)

    setTimeout(() => {
      el.remove()
    }, 3500)
  }

  return {
    toast
  }
}

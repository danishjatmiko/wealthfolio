// One shared <audio> element for every clip in the game. Mobile Safari
// only lets a page play audio from an element that was first started by
// a tap; reusing the same element afterwards (e.g. the quiz reading the
// next question on its own) keeps working, while a fresh `new Audio()`
// each time would get blocked.
let el: HTMLAudioElement | null = null
let token = 0

function audio(): HTMLAudioElement {
  if (!el) {
    el = new Audio()
    el.preload = 'auto'
  }
  return el
}

const sleep = (ms: number) => new Promise<void>((r) => setTimeout(r, ms))

function playOne(a: HTMLAudioElement, url: string, mine: number): Promise<void> {
  return new Promise((resolve) => {
    let iv = 0
    const done = () => {
      a.removeEventListener('ended', done)
      a.removeEventListener('error', done)
      window.clearInterval(iv)
      resolve()
    }
    a.addEventListener('ended', done)
    a.addEventListener('error', done)
    // stop() pauses the element, which never fires `ended` — poll so an
    // interrupted sequence still unwinds.
    iv = window.setInterval(() => {
      if (mine !== token) done()
    }, 100)
    a.src = url
    a.play().catch(done)
  })
}

/** Stops whatever is playing and cancels any running sequence. */
export function stopAudio() {
  token++
  el?.pause()
}

/**
 * Plays clips back to back, interrupting anything already playing. A
 * `null` entry is a short pause. `onStep` fires as each entry starts.
 * Resolves true if it ran to the end, false if something else cut in.
 */
export async function playSequence(
  urls: (string | null)[],
  onStep?: (i: number) => void,
  gapMs = 0,
): Promise<boolean> {
  stopAudio()
  const mine = token
  const a = audio()
  for (let i = 0; i < urls.length; i++) {
    if (mine !== token) return false
    onStep?.(i)
    const url = urls[i]
    if (url) await playOne(a, url, mine)
    else await sleep(350)
    if (gapMs && i < urls.length - 1) await sleep(gapMs)
  }
  return mine === token
}

export const playClip = (url: string) => playSequence([url])

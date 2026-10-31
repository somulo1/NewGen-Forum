// How long to wait before the next try after a dropped connection or a failed request.
// "Full jitter": the wait is a random time between a small floor and a ceiling that doubles with each failed try (1 s, 2 s, 4 s ...
// up to 30 s). When a server restarts and a million browsers lose their connection together, a fixed 1-2-4 schedule would bring them
// all back in the same second; a random spread brings them back evenly over the whole window, so the server is never hit all at once.
export const FLOOR_MS = 300;
export const CAP_MS = 30000;
export const BASE_MS = 1000;

export function ceiling(attempt) { return Math.min(CAP_MS, BASE_MS * 2 ** Math.max(0, attempt - 1)); }

// attempt: 1 for the first retry. rand: a function giving a number in [0, 1) (Math.random; replaceable so it can be tested).
export function retryDelay(attempt, rand = Math.random) {
  const top = ceiling(attempt);
  return Math.round(FLOOR_MS + rand() * (top - FLOOR_MS));
}

// A short random pause before work that every browser starts at the same moment (for example when a tab becomes visible again).
export function spread(maxMs, rand = Math.random) { return Math.round(rand() * maxMs); }

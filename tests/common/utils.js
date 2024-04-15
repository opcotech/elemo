export function randomIntBetween(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

export function deepEqual(lhs, rhs) {
  const keys1 = Object.keys(lhs);
  const keys2 = Object.keys(rhs);
  if (keys1.length !== keys2.length) {
    return false;
  }
  for (const key of keys1) {
    const val1 = lhs[key];
    const val2 = rhs[key];
    const areObjects = isObject(val1) && isObject(val2);
    if (
      areObjects && !deepEqual(val1, val2) ||
      !areObjects && val1 !== val2
    ) {
      return false;
    }
  }
  return true;
}

function isObject(obj) {
  return obj != null && typeof obj === 'object';
}

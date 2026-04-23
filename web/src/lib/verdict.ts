export function verdictTone(verdict: string | undefined) {
  switch (verdict) {
    case 'AC':
      return 'success';
    case 'WA':
    case 'CE':
    case 'RE':
    case 'TLE':
      return 'danger';
    case 'UNJUDGEABLE':
      return 'warning';
    default:
      return 'neutral';
  }
}

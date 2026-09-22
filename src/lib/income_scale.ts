export function incomeTimelineDomain(
  points: Array<Record<string, unknown>>,
  visibleGroups: string[]
): [number, number] {
  const domain = points.reduce<[number, number]>(
    ([min, max], point) => {
      let negative = 0;
      let positive = 0;

      visibleGroups.forEach((group) => {
        const value = Number(point[group] ?? 0);
        if (value < 0) {
          negative += value;
        } else {
          positive += value;
        }
      });

      return [Math.min(min, negative), Math.max(max, positive)];
    },
    [0, 0]
  );

  return domain[0] === domain[1] ? [0, 1] : domain;
}

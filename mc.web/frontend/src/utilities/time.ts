/** Convert days to nanoseconds for Go time.Duration */
export function DaysToNanoseconds(days: number): number {
    return Math.round(days * 24 * 60 * 60 * 1e9);
}

export function NanosecondsToDays(nanoseconds: number): number {
    return Math.round(nanoseconds / (24 * 60 * 60 * 1e9));
}
### TODO:

- no re-assignment to constants
- \_ for unused values

### Objective

fibonacci :: (n: number) => number {
if n == 0 { return 0 }
if n == 1 { return 1 }

    (prev, cur) := (1, 1)
    for _ := 3..=n {
        (prev, cur) = (cur, prev + cur)
    }
    return cur

}

#[derive(Debug, Clone, Copy, Default)]
pub struct RunningStats {
    count: u64,
    mean: f64,
    squared_deviation: f64,
}

impl RunningStats {
    pub fn observe(&mut self, value: f64) {
        self.count += 1;
        let delta = value - self.mean;
        self.mean += delta / self.count as f64;
        let adjusted = value - self.mean;
        self.squared_deviation += delta * adjusted;
    }

    pub fn merge(&mut self, other: Self) {
        if other.count == 0 {
            return;
        }
        if self.count == 0 {
            *self = other;
            return;
        }

        let combined = self.count + other.count;
        let delta = other.mean - self.mean;
        self.squared_deviation += other.squared_deviation
            + delta * delta * self.count as f64 * other.count as f64 / combined as f64;
        self.mean += delta * other.count as f64 / combined as f64;
        self.count = combined;
    }

    pub fn mean(&self) -> Option<f64> {
        (self.count > 0).then_some(self.mean)
    }

    pub fn sample_variance(&self) -> Option<f64> {
        (self.count > 1).then_some(self.squared_deviation / (self.count - 1) as f64)
    }
}

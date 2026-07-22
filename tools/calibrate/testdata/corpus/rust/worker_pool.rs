use std::sync::{mpsc, Arc, Mutex};
use std::thread;

#[derive(Debug)]
pub struct Job {
    pub id: u64,
    pub payload: String,
}

#[derive(Debug, PartialEq, Eq)]
pub struct JobResult { pub id: u64, pub checksum: u64 }

pub fn run(jobs: Vec<Job>, workers: usize) -> Result<Vec<JobResult>, &'static str> {
    if workers == 0 {
        return Err("workers must be positive");
    }

    let (job_tx, job_rx) = mpsc::channel::<Job>();
    let (result_tx, result_rx) = mpsc::channel::<JobResult>();
    let shared_rx = Arc::new(Mutex::new(job_rx));
    let mut handles = Vec::new();

    for _ in 0..workers {
        let receiver = Arc::clone(&shared_rx);
        let sender = result_tx.clone();
        handles.push(thread::spawn(move || loop {
            let next = receiver.lock().unwrap().recv();
            let Ok(job) = next else { break };
            let checksum = job.payload.bytes().fold(0_u64, |sum, byte| {
                sum.wrapping_mul(31).wrapping_add(byte as u64)
            });
            if sender.send(JobResult { id: job.id, checksum }).is_err() {
                break;
            }
        }));
    }
    drop(result_tx);

    let expected = jobs.len();
    for job in jobs {
        job_tx.send(job).map_err(|_| "queue closed")?;
    }
    drop(job_tx);

    let mut results: Vec<_> = result_rx.iter().take(expected).collect();
    for handle in handles {
        handle.join().map_err(|_| "panic")?;
    }
    results.sort_by_key(|result| result.id);
    Ok(results)
}

package nachos.threads;

import nachos.machine.*;

import java.util.LinkedList;

/**
 * An implementation of condition variables that disables interrupt()s for
 * synchronization.
 *
 * <p>
 * You must implement this.
 *
 * @see	nachos.threads.Condition
 */
public class Condition2 {
  /**
   * Allocate a new condition variable.
   *
   * @param	conditionLock	the lock associated with this condition
   *				variable. The current thread must hold this
   *				lock whenever it uses <tt>sleep()</tt>,
   *				<tt>wake()</tt>, or <tt>wakeAll()</tt>.
   */
  public Condition2(Lock conditionLock) {
    this.conditionLock = conditionLock;
    waitQueue = new LinkedList<KThread>();
  }

  /**
   * Atomically release the associated lock and go to sleep on this condition
   * variable until another thread wakes it using <tt>wake()</tt>. The
   * current thread must hold the associated lock. The thread will
   * automatically reacquire the lock before <tt>sleep()</tt> returns.
   */
  public void sleep() {
    Lib.assertTrue(conditionLock.isHeldByCurrentThread());

    waitQueue.add(KThread.currentThread());

    conditionLock.release();

    boolean intStatus = Machine.interrupt().disable();

    KThread.sleep();

    Machine.interrupt().restore(intStatus);

    conditionLock.acquire();
  }

  /**
   * Wake up at most one thread sleeping on this condition variable. The
   * current thread must hold the associated lock.
   */
  public void wake() {
    Lib.assertTrue(conditionLock.isHeldByCurrentThread());

    if (!waitQueue.isEmpty()) {
      boolean intStatus = Machine.interrupt().disable();

      waitQueue.removeFirst().ready();

      Machine.interrupt().restore(intStatus);
    }
  }

  /**
   * Wake up all threads sleeping on this condition variable. The current
   * thread must hold the associated lock.
   */
  public void wakeAll() {
    Lib.assertTrue(conditionLock.isHeldByCurrentThread());

    while (!waitQueue.isEmpty())
      wake();
  }

  private Lock conditionLock;
  private LinkedList<KThread> waitQueue;

  /**
   * Tests for the Condition2.
   */
  public static void waitTest() {
    Lock lock = new Lock();
    Condition2 impossible = new Condition2(lock);
    lock.acquire();
    System.out.println("*** Acquire lock for impossible.");
    impossible.sleep();
  }

  public static void joinTest() {
    Lock lock = new Lock();
    Condition2 sonFinished = new Condition2(lock);
    lock.acquire();

    KThread sonThread = new KThread(new Runnable() {
      public void run() {
        System.out.println("*** Son thread finished.");
        lock.acquire();
        sonFinished.wake();
        lock.release();
      }
    });

    sonThread.fork();
    sonFinished.sleep();
    System.out.println("*** Son thread joined.");
  }

  public static void selfTest() {
    joinTest();
  }
}

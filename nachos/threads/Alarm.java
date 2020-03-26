package nachos.threads;

import nachos.machine.*;

import java.util.LinkedList;
import java.util.Iterator;

/**
 * Uses the hardware timer to provide preemption, and to allow threads to sleep
 * until a certain time.
 */
public class Alarm {
  /**
   * Allocate a new Alarm. Set the machine's timer interrupt handler to this
   * alarm's callback.
   *
   * <p><b>Note</b>: Nachos will not function correctly with more than one
   * alarm.
   */
  public Alarm() {
    Machine.timer().setInterruptHandler(new Runnable() {
      public void run() { timerInterrupt(); }
    });

    waitQueue = new LinkedList<Pair>();
  }

  /**
   * The timer interrupt handler. This is called by the machine's timer
   * periodically (approximately every 500 clock ticks). Causes the current
   * thread to yield, forcing a context switch if there is another thread
   * that should be run.
   */
  public void timerInterrupt() {
    long currentTime = Machine.timer().getTime();

    for (Iterator it = waitQueue.iterator(); it.hasNext();) {
      Pair waitingObj = (Pair) it.next();
      if (waitingObj.getWaitTime() <= currentTime) {
        waitingObj.getThread().ready();
        it.remove();
      }
    }

    KThread.currentThread().yield();
  }

  /**
   * Put the current thread to sleep for at least <i>x</i> ticks,
   * waking it up in the timer interrupt handler. The thread must be
   * woken up (placed in the scheduler ready set) during the first timer
   * interrupt where
   *
   * <p><blockquote>
   * (current time) >= (WaitUntil called time)+(x)
   * </blockquote>
   *
   * @param	x	the minimum number of clock ticks to wait.
   *
   * @see	nachos.machine.Timer#getTime()
   */
  public void waitUntil(long x) {
    // for now, cheat just to get something working (busy waiting is bad)
    long wakeTime = Machine.timer().getTime() + x;

    if (wakeTime > Machine.timer().getTime()) {
      waitQueue.add( new Pair(KThread.currentThread(), wakeTime) );

      boolean intStatus = Machine.interrupt().disable();

      KThread.sleep();

      Machine.interrupt().restore(intStatus);
    }
  }

  protected class Pair {
    Pair(KThread thread, long waitTime) {
      this.thread = thread;
      this.waitTime = waitTime;
    }

    KThread getThread() { return this.thread; }
    long getWaitTime() { return this.waitTime; }

    private KThread thread;
    private long waitTime;
  }

  private LinkedList<Pair> waitQueue;

  /**
   * Tests for Alarm.
   */
  private static class clockTest implements Runnable {
    clockTest(long sleepTime) { this.sleepTime = sleepTime; }

    public void run() {
      ThreadedKernel.alarm.waitUntil(sleepTime);
      System.out.println("*** thread " + KThread.currentThread()
                         + " finished.");
    }

    private long sleepTime;
  }

  public static void clocksTest() {
    LinkedList<KThread> clocksQueue = new LinkedList<KThread>();
    for (int i = 1; i <= 5; i++){
      KThread newClock = new KThread(new clockTest(i * 1000000));
      clocksQueue.add(newClock);
      newClock.setName("clock " + i);
      newClock.fork();
    }

    while (!clocksQueue.isEmpty()){
      clocksQueue.removeFirst().join();
    }
  }

  public static void selfTest() {
    clocksTest();
  }
}

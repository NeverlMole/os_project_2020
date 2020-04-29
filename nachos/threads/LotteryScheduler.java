package nachos.threads;

import nachos.machine.*;

import java.util.LinkedList;
import java.util.Iterator;
import java.util.Random;

/**
 * A scheduler that chooses threads using a lottery.
 *
 * <p>
 * A lottery scheduler associates a number of tickets with each thread. When a
 * thread needs to be dequeued, a random lottery is held, among all the tickets
 * of all the threads waiting to be dequeued. The thread that holds the winning
 * ticket is chosen.
 *
 * <p>
 * Note that a lottery scheduler must be able to handle a lot of tickets
 * (sometimes billions), so it is not acceptable to maintain state for every
 * ticket.
 *
 * <p>
 * A lottery scheduler must partially solve the priority inversion problem; in
 * particular, tickets must be transferred through locks, and through joins.
 * Unlike a priority scheduler, these tickets add (as opposed to just taking
 * the maximum).
 */
public class LotteryScheduler extends PriorityScheduler {
  /**
   * Allocate a new lottery scheduler.
   */
 public LotteryScheduler() {}
 
  /**
   * Allocate a new lottery thread queue.
   *
   * @param	transferPriority	<tt>true</tt> if this queue should
   *					transfer tickets from waiting threads
   *					to the owning thread.
   * @return	a new lottery thread queue.
   */
  @Override
  public ThreadQueue newThreadQueue(boolean transferPriority) {
    return new LotteryQueue(transferPriority);
  }

  @Override
  protected newThreadState getThreadState(KThread thread) {
    if (thread.schedulingState == null)
      thread.schedulingState = new newThreadState(thread);

    return (newThreadState)thread.schedulingState;
  }
  /*************************************************************************************************/
  /**
   * A <tt>ThreadQueue</tt> that draws threads through lottery.
   */
  protected class LotteryQueue extends PriorityQueue {
    LotteryQueue(boolean transferPriority) {
      super(transferPriority);
      ran = new Random();
    }

    @Override
    public void waitForAccess(KThread thread) {
      Lib.assertTrue(Machine.interrupt().disabled());
      OrderedKThread temp = new OrderedKThread(thread, 0, this);
      waitPQueue.add(temp);
      getThreadState(thread).waitForAccess(temp);
      CheckAcquire();
    }

    @Override
    public void acquire(KThread thread) {
      Lib.assertTrue(Machine.interrupt().disabled());
      newThreadState threadstate = getThreadState(thread);
      AcquireList.add(threadstate.acquire(this));
      if (!waitPQueue.isEmpty())
        threadstate.updateLottery();
    }

    @Override
    public KThread nextThread() {
      Lib.assertTrue(Machine.interrupt().disabled());
      if (waitPQueue.isEmpty())
        return null;
      OrderedKThread nextt = poll2();
      getThreadState(nextt.thread).RemoveWait(nextt);
      /*Here we remove all the acquired threads, which can be modified for other
       * application*/
      for (Iterator i = AcquireList.iterator(); i.hasNext();) {
        OrderedKThread iter = (OrderedKThread)(i.next());
        getThreadState(iter.thread).RemoveAcquire(iter);
      }
      AcquireList = new LinkedList<OrderedKThread>();

      acquire(nextt.thread);
      return (KThread)(nextt.thread);
    }

    public OrderedKThread poll2() {
      int Ticketsum = SumTicket();
      int Lottery = ran.nextInt(Ticketsum) + 1;
      for (Iterator i = waitPQueue.iterator(); i.hasNext();) {
        OrderedKThread iter = (OrderedKThread)(i.next());
        Lottery = Lottery - getThreadState(iter.thread).Ticket;
        if(Lottery<=0){
          waitPQueue.remove(iter);
          return iter;
        }
      }
      return null;
    }

    /* calculate the sum of tickets in the queue*/
    public int SumTicket() {
      int SumTick = 0;
      for (Iterator i = waitPQueue.iterator(); i.hasNext();) {
        OrderedKThread iter = (OrderedKThread)(i.next());
        SumTick = SumTick + getThreadState(iter.thread).Ticket;
      }
      return SumTick;
    }

    /* check if there are acquired threads in need of ticket transfer*/
    public void CheckAcquire() {
      for (Iterator i = AcquireList.iterator(); i.hasNext();) {
        OrderedKThread iter = (OrderedKThread)(i.next());
        newThreadState iterstate = getThreadState(iter.thread);
        iterstate.updateLottery();
      }
    }

    private Random ran;
  }



  /*************************************************************************************************/

  protected class newThreadState extends ThreadState {

    public newThreadState(KThread thread) {
      super(thread);
      this.thread = thread;
      setPriority(priorityDefault);
    }
 
    @Override
    public void setPriority(int priority) {
      Lib.assertTrue(Machine.interrupt().disabled());
      this.priority = priority;
      updateLottery();
    }

    public int getTicket() {
      return Ticket;
    }

    public int updateLottery() {
      Lib.assertTrue(Machine.interrupt().disabled());
      int PrevTicket = Ticket;
      Ticket = priority;
      if(Ticket<1) Ticket = 1;

      for (Iterator i = AcquireList.iterator(); i.hasNext();) {
        OrderedKThread iter = (OrderedKThread)(i.next());

        if (((LotteryQueue)(iter.queue)).transferPriority) {
          Ticket = Ticket + ((LotteryQueue)(iter.queue)).SumTicket();
        }
      }

      if (Ticket != PrevTicket) {
        for (Iterator i = WaitList.iterator(); i.hasNext();) {
          OrderedKThread iter = (OrderedKThread)(i.next());
          if (((LotteryQueue)(iter.queue)).transferPriority) {
            ((LotteryQueue)(iter.queue)).CheckAcquire();
          }
        }
      }

      return Ticket;
    }

    @Override
    public void RemoveAcquire(OrderedKThread othread) {
      AcquireList.remove(othread);
      updateLottery();
    }

    protected int Ticket;
  }
}

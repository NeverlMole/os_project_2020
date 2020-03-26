package nachos.threads;

import nachos.machine.*;

import java.util.Iterator;
import java.util.LinkedList;
import java.util.Comparator;

/**
 * A scheduler that chooses threads based on their priorities.
 *
 * <p>
 * A priority scheduler associates a priority with each thread. The next thread
 * to be dequeued is always a thread with priority no less than any other
 * waiting thread's priority. Like a round-robin scheduler, the thread that is
 * dequeued is, among all the threads of the same (highest) priority, the
 * thread that has been waiting longest.
 *
 * <p>
 * Essentially, a priority scheduler gives access in a round-robin fassion to
 * all the highest-priority threads, and ignores all other threads. This has
 * the potential to
 * starve a thread if there's always a thread waiting with higher priority.
 *
 * <p>
 * A priority scheduler must partially solve the priority inversion problem; in
 * particular, priority must be donated through locks, and through joins.
 */
public class PriorityScheduler extends Scheduler {
  /**
   * Allocate a new priority scheduler.
   */
  public PriorityScheduler() {}

  /**
   * Allocate a new priority thread queue.
   *
   * @param	transferPriority	<tt>true</tt> if this queue should
   *					transfer priority from waiting threads
   *					to the owning thread.
   * @return	a new priority thread queue.
   */
  public ThreadQueue newThreadQueue(boolean transferPriority) {
    return new PriorityQueue(transferPriority);
  }


  public int getPriority(KThread thread) {
    Lib.assertTrue(Machine.interrupt().disabled());

    return getThreadState(thread).getPriority();
  }

  public int getEffectivePriority(KThread thread) {
    Lib.assertTrue(Machine.interrupt().disabled());

    return getThreadState(thread).getEffectivePriority();
  }

  public void setPriority(KThread thread, int priority) {
    Lib.assertTrue(Machine.interrupt().disabled());

    Lib.assertTrue(priority >= priorityMinimum && priority <= priorityMaximum);

    getThreadState(thread).setPriority(priority);
  }

  public boolean increasePriority() {
    boolean intStatus = Machine.interrupt().disable();

    KThread thread = KThread.currentThread();

    int priority = getPriority(thread);
    if (priority == priorityMaximum)
      return false;

    setPriority(thread, priority + 1);

    Machine.interrupt().restore(intStatus);
    return true;
  }

  public boolean decreasePriority() {
    boolean intStatus = Machine.interrupt().disable();

    KThread thread = KThread.currentThread();

    int priority = getPriority(thread);
    if (priority == priorityMinimum)
      return false;

    setPriority(thread, priority - 1);

    Machine.interrupt().restore(intStatus);
    return true;
  }

  /**
   * The default priority for a new thread. Do not change this value.
   */
  public static final int priorityDefault = 1;
  /**
   * The minimum priority that a thread can have. Do not change this value.
   */
  public static final int priorityMinimum = 0;
  /**
   * The maximum priority that a thread can have. Do not change this value.
   */
  public static final int priorityMaximum = 7;

  /**
   * Return the scheduling state of the specified thread.
   *
   * @param	thread	the thread whose scheduling state to return.
   * @return	the scheduling state of the specified thread.
   */
  protected ThreadState getThreadState(KThread thread) {
    if (thread.schedulingState == null)
      thread.schedulingState = new ThreadState(thread);

    return (ThreadState)thread.schedulingState;
  }

  /** An entry in queues to record thread registration
   *  order: the order in the waitqueue to ensure FIFO
   * 
   *  used in both resource and thread records.
   */ 	
  private class OrderedKThread {
     public KThread thread;
     public int order;
     public ThreadQueue queue;
      OrderedKThread(KThread thread, int order, ThreadQueue queue){
	this.thread = thread;
        this.order = order;
        this.queue = queue;
      }	
    }

  public static int max(int a,int b){
    if(a>b) return a; else return b;
  }
  /*************************************************************************************************/
  /**
   * A <tt>ThreadQueue</tt> that sorts threads by priority.
   */
  protected class PriorityQueue extends ThreadQueue {
    PriorityQueue(boolean transferPriority) {
      this.transferPriority = transferPriority;
    }

    public void waitForAccess(KThread thread) {
      Lib.assertTrue(Machine.interrupt().disabled());
      OrderedKThread temp = new OrderedKThread(thread, waitPQueue.size(), this);
      waitPQueue.add(temp);
      //print();
      ThreadState threadstate = getThreadState(thread);
      threadstate.waitForAccess(temp);
      /* check donations: */
      CheckAcquire(threadstate.CEpriority);
    }

    public void acquire(KThread thread) {
      Lib.assertTrue(Machine.interrupt().disabled());
      ThreadState threadstate = getThreadState(thread);
      AcquireList.add(threadstate.acquire(this));
      /* update the effective priority if there are waiting threads*/
      if(!waitPQueue.isEmpty()) threadstate.getEffectivePriority();
    }

    public KThread nextThread() {
      Lib.assertTrue(Machine.interrupt().disabled());
      if (waitPQueue.isEmpty())
	   return null;
      OrderedKThread nextt = poll();
      getThreadState(nextt.thread).RemoveWait(nextt);  
      /*Here we remove all the acquired threads, which can be modified for other application*/
      for (Iterator i=AcquireList.iterator(); i.hasNext(); ){
         OrderedKThread iter = (OrderedKThread) (i.next());
         getThreadState(iter.thread).RemoveAcquire(iter);
      }
      AcquireList = new LinkedList<OrderedKThread>();
      return (KThread) (nextt.thread);
    }

    /*Comparator for the priority queue*/
    public int compare(OrderedKThread t1, OrderedKThread t2){
      int Cmpr = getThreadState(t2.thread).CEpriority-getThreadState(t1.thread).CEpriority;
      if (Cmpr==0) return t1.order-t2.order;      
      return Cmpr;
    };

    public OrderedKThread poll(){
      OrderedKThread poll = null;
      for (Iterator i=waitPQueue.iterator(); i.hasNext(); ){
        OrderedKThread iter = (OrderedKThread) (i.next());
        if (poll==null) poll=iter; else {
          if(compare(poll, iter)>0) poll=iter;
        } 
      }
      for (Iterator i=waitPQueue.iterator(); i.hasNext(); ){
        OrderedKThread iter = (OrderedKThread) (i.next());
	if (iter.order>poll.order) iter.order--;
      }
      waitPQueue.remove(poll);
      return poll;
    }

    public boolean isEmpty(){
       return waitPQueue.isEmpty();
    }

    /* Returns the highest effective priority in the waiting queue */
    public int MaxPriority(){
      int Maxprior = priorityMinimum;
      for (Iterator i=waitPQueue.iterator(); i.hasNext(); ){
         OrderedKThread iter = (OrderedKThread) (i.next());
	 Maxprior = max(Maxprior, getThreadState(iter.thread).CEpriority); 
      }
      return Maxprior;
    }

    /* check if there are acquired threads in need of priority donation*/
    public void CheckAcquire(int Cpriority){
      for (Iterator i=AcquireList.iterator(); i.hasNext(); ){
         OrderedKThread iter = (OrderedKThread) (i.next());
         ThreadState iterstate = getThreadState(iter.thread);
	 if(Cpriority>iterstate.CEpriority) iterstate.getEffectivePriority();
      }
    }
    public void print() {
      Lib.assertTrue(Machine.interrupt().disabled());	 
      System.out.print("Queue size: ");
      System.out.print(waitPQueue.size()); 
      System.out.print(", with: ");
      for (Iterator i=waitPQueue.iterator(); i.hasNext(); ){
          KThread iter = (KThread) ((OrderedKThread) i.next()).thread ;
	  System.out.print(iter + " effective priority: " + getThreadState(iter).CEpriority + "\n"); 
      }
      System.out.print("\n");
    }
    /**
     * <tt>true</tt> if this queue should transfer priority from waiting
     * threads to the owning thread.
     */
    public boolean transferPriority;
    private LinkedList<OrderedKThread> waitPQueue = new LinkedList<OrderedKThread>();
    private LinkedList<OrderedKThread> AcquireList = new LinkedList<OrderedKThread>();
  }
	
  /*************************************************************************************************/
  /**
   * The scheduling state of a thread. This should include the thread's
   * priority, its effective priority, any objects it owns, and the queue
   * it's waiting for, if any.
   *
   * @see	nachos.threads.KThread#schedulingState
   */
  protected class ThreadState {
    /**
     * Allocate a new <tt>ThreadState</tt> object and associate it with the
     * specified thread.
     *
     * @param	thread	the thread this state belongs to.
     */
    public ThreadState(KThread thread) {
      this.thread = thread;
      setPriority(priorityDefault);
    }

    /**
     * Return the priority of the associated thread.
     *
     * @return	the priority of the associated thread.
     */
    public int getPriority() { return priority; }

    /**
     * Return the effective priority of the associated thread.
     *
     * @return	the effective priority of the associated thread.
     */
    public int getEffectivePriority() {	
      Lib.assertTrue(Machine.interrupt().disabled());
      int PrevCEpriority = CEpriority;	 
      CEpriority = priority;
      for (Iterator i=AcquireList.iterator(); i.hasNext(); ){
        OrderedKThread iter = (OrderedKThread) (i.next());
        CEpriority = max(CEpriority, ((PriorityQueue) (iter.queue)).MaxPriority());
      }
      /* the rise of effective priority may affect the donation in other waiting queues*/
      if(CEpriority>PrevCEpriority) {
        for (Iterator i=WaitList.iterator(); i.hasNext(); ){
          OrderedKThread iter = (OrderedKThread) (i.next());
          ((PriorityQueue) (iter.queue)).CheckAcquire(CEpriority);
        }
      }
      return CEpriority;
    }

    /**
     * Set the priority of the associated thread to the specified value.
     *
     * @param	priority	the new priority.
     */
    public void setPriority(int priority) {  
      Lib.assertTrue(Machine.interrupt().disabled());	  
      this.priority = priority;
      getEffectivePriority();
    }

    /**
     * Called when <tt>waitForAccess(thread)</tt> (where <tt>thread</tt> is
     * the associated thread) is invoked on the specified priority queue.
     * The associated thread is therefore waiting for access to the
     * resource guarded by <tt>waitQueue</tt>. This method is only called
     * if the associated thread cannot immediately obtain access.
     *
     * @param	waitQueue	the queue that the associated thread is
     *				now waiting on.
     *
     * @see	nachos.threads.ThreadQueue#waitForAccess
     */
    public void waitForAccess(OrderedKThread othread) {
      WaitList.add(othread);
    }

    public void RemoveWait(OrderedKThread othread){
      WaitList.remove(othread);
    }
    
    public void RemoveAcquire(OrderedKThread othread){
      AcquireList.remove(othread);
      getEffectivePriority();
    }
    /**
     * Called when the associated thread has acquired access to whatever is
     * guarded by <tt>waitQueue</tt>. This can occur either as a result of
     * <tt>acquire(thread)</tt> being invoked on <tt>waitQueue</tt> (where
     * <tt>thread</tt> is the associated thread), or as a result of
     * <tt>nextThread()</tt> being invoked on <tt>waitQueue</tt>.
     *
     * @see	nachos.threads.ThreadQueue#acquire
     * @see	nachos.threads.ThreadQueue#nextThread
     */
    public OrderedKThread acquire(PriorityQueue waitQueue) {
      OrderedKThread temp = new OrderedKThread(thread, 0, waitQueue);
      AcquireList.add(temp);
      return temp;
    }
     
    /** The thread with which this object is associated. */
    protected KThread thread;
    /** The priority of the associated thread. */
    protected int priority;
    /* Cached effective priority*/
    protected int CEpriority=priorityDefault;
    private LinkedList<OrderedKThread> WaitList = new LinkedList<OrderedKThread>();
    private LinkedList<OrderedKThread> AcquireList = new LinkedList<OrderedKThread>();
  }
}

package nachos.threads;

import nachos.machine.*;

import java.util.LinkedList;

/**
 * A <i>communicator</i> allows threads to synchronously exchange 32-bit
 * messages. Multiple threads can be waiting to <i>speak</i>,
 * and multiple threads can be waiting to <i>listen</i>. But there should never
 * be a time when both a speaker and a listener are waiting, because the two
 * threads can be paired off at this point.
 */
public class Communicator {
  /**
   * Allocate a new communicator.
   */
  public Communicator() {
    lock = new Lock();
    listenChannel = new Condition(lock);
    speakChannel = new Condition(lock);
    speakerSended = new Condition(lock);
    listenerReceived = new Condition(lock);
  }

  /**
   * Wait for a thread to listen through this communicator, and then transfer
   * <i>word</i> to the listener.
   *
   * <p>
   * Does not return until this thread is paired up with a listening thread.
   * Exactly one listener should receive <i>word</i>.
   *
   * @param	word	the integer to transfer.
   */
  public void speak(int word) {
    Lib.assertTrue(numSpeakers >= 0);

    lock.acquire();

    Lib.debug(dbgCommunicator, "Thread " + KThread.currentThread()
                        + " starts speak with " + numSpeakers
                        + " other speakers.");

    numSpeakers += 1;

    if (numSpeakers != 1) {
      speakChannel.sleep();
    }

    Lib.assertTrue(numSpeakersChannel == 0);
    Lib.debug(dbgCommunicator, "Thread " + KThread.currentThread()
                        + " enters speak channel.");

    numSpeakersChannel += 1;       // Here the speaker enter the channel.

    messageSended = true;
    buf = word;

    speakerSended.wake();
    listenerReceived.sleep();

    speakChannel.wake();

    numSpeakersChannel -= 1;       // Here the speaker leave the channel.

    numSpeakers -= 1;
    lock.release();
  }

  /**
   * Wait for a thread to speak through this communicator, and then return
   * the <i>word</i> that thread passed to <tt>speak()</tt>.
   *
   * @return	the integer transferred.
   */
  public int listen() {
    Lib.assertTrue(numListeners >= 0);

    lock.acquire();

    Lib.debug(dbgCommunicator, "Thread " + KThread.currentThread()
                        + " starts listen with " + numListeners
                        + " other listeners.");

    numListeners += 1;

    if (numListeners != 1) {
      listenChannel.sleep();
    }

    Lib.assertTrue(numListenersChannel == 0);

    Lib.debug(dbgCommunicator, "Thread " + KThread.currentThread()
                        + " enters listen channel.");

    numListenersChannel += 1;        // Here the listener enters the channel.

    while (!messageSended) {
      speakerSended.sleep();
    }

    int returnValue = buf;
    messageSended = false;
    listenerReceived.wake();

    listenChannel.wake();

    numListenersChannel -= 1;     // Here the listener leavers the channel.

    numListeners -= 1;
    lock.release();
    return returnValue;
  }

  private Lock lock;
  private Condition listenChannel;
  private Condition speakChannel;
  private Condition speakerSended;
  private Condition listenerReceived;

  private int numSpeakers = 0;
  private int numListeners = 0;
  private boolean messageSended = false;
  private int numListenersChannel = 0;
  private int numSpeakersChannel = 0;
  private int buf;

  private static final char dbgCommunicator = 'c';

  /**
   * Tests cases for Communicator.
   */

  public static class SpeakerTest implements Runnable {
    SpeakerTest(int word, Communicator com) {
      this.word = word;
      this.com = com;
    }

    public void run() {
      com.speak(word);
      System.out.println("Speaker " + KThread.currentThread() + " sended "
                          + word + ".");
    }

    private int word;
    private Communicator com;
  }

  public static class ListenerTest implements Runnable {
    ListenerTest(Communicator com) {
      this.com = com;
    }

    public void run() {
      int getWord = com.listen();
      System.out.println("Listener " + KThread.currentThread() + " received "
                          + getWord + ".");
    }

    private Communicator com;
  }

  public static void listenerWaitingTest() {
    Communicator com = new Communicator();

    for (int i = 1; i <= 5; i++) {
      KThread listener = new KThread(new ListenerTest(com));
      listener.setName("(" + i + ")");
      listener.fork();
    }

    KThread.yield();

    for (int i = 1; i <= 5; i++) {
      (new SpeakerTest(i, com)).run();
    }
  }

  public static void speakerWaitingTest() {
    Communicator com = new Communicator();

    for (int i = 1; i <= 5; i++) {
      KThread speaker = new KThread(new SpeakerTest(i, com));
      speaker.setName("(" + i + ")");
      speaker.fork();
    }

    KThread.yield();

    for (int i = 1; i <= 5; i++) {
      (new ListenerTest(com)).run();
    }
  }

  public static void generalTest() {
    Communicator com = new Communicator();
    LinkedList<KThread> speakers = new LinkedList<KThread>();
    LinkedList<KThread> listeners = new LinkedList<KThread>();

    for (int i = 1; i <= 5; i++) {
      KThread speaker = new KThread(new SpeakerTest(i, com));
      speaker.setName("(" + i + ")");
      speakers.add(speaker);
    }

    for (int i = 1; i <= 5; i++) {
      KThread listener = new KThread(new ListenerTest(com));
      listener.setName("(" + i + ")");
      listeners.add(listener);
    }

    speakers.get(0).fork();
    speakers.get(1).fork();
    speakers.get(2).fork();
    listeners.get(0).fork();
    speakers.get(3).fork();
    listeners.get(1).fork();
    listeners.get(2).fork();
    listeners.get(3).fork();
    listeners.get(4).fork();
    speakers.get(4).fork();

    speakers.get(4).join();
  }

  public static void selfTest() {
    speakerWaitingTest();
    listenerWaitingTest();
    generalTest();
  }
}

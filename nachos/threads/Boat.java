package nachos.threads;
import nachos.ag.BoatGrader;

public class Boat {
  static BoatGrader bg;

  static KThread parent_thread;
  static int child_O_num;
  static int child_M_num;
  static int adult_O_num;
  static int adult_M_num;
  static boolean has_pilot;
  static int boat;
  static Lock lock;
  static Condition con_child_O;
  static Condition con_child_M;
  static Condition con_adult;
  static Condition finish;

  public static void selfTest() {
    BoatGrader b = new BoatGrader();
	
    begin(3, 3, b);

    begin(10, 9, b);

    begin(17, 5, b);

    begin(5, 20, b);

    begin(0, 2, b);

    begin(20, 2, b);
  }

  public static void begin(int adults, int children, BoatGrader b) {
    // Store the externally generated autograder in a class
    // variable to be accessible by children.
    bg = b;
	
	parent_thread = KThread.currentThread();
	child_O_num = children;
	adult_O_num = adults;
	child_M_num = 0;
	adult_M_num = 0;
	has_pilot = false;
	boat = 0; 
	//boat 0 in O, 1 in M
	lock = new Lock();
	con_adult = new Condition(lock); //Condition 2
	con_child_O = new Condition(lock); //Condition 1
	con_child_M = new Condition(lock); //Condition 3
	finish = new Condition(lock);

    // Instantiate global variables here
	
	lock.acquire();

	for (int i = 1; i <= adults; i++){
		Runnable r = new Runnable() { 
			public void run() { AdultItinerary(); }
  		};
		KThread t = new KThread(r);
		t.setName("Adult Thread " + i);
		t.fork();
	}

 	for (int i = 1; i <= children; i++){
		Runnable r = new Runnable() { 
			public void run() { ChildItinerary(); }
  		};
		KThread t = new KThread(r);
		t.setName("Child Thread " + i);
		t.fork();
	}			
	
	finish.sleep();

	lock.release();
	 // Create threads here. See section 3.4 of the Nachos for Java
    // Walkthrough linked from the projects page.

  }

  static void AdultItinerary() {
    bg.initializeAdult(); // Required for autograder interface. Must be the
                          // first thing called.
    // DO NOT PUT ANYTHING ABOVE THIS LINE.

    /* This is where you should put your solutions. Make calls
       to the BoatGrader to show that it is synchronized. For
       example:
           bg.AdultRowToMolokai();
       indicates that an adult has rowed the boat across to Molokai
    */
	
	lock.acquire();

	while (boat == 1 || child_O_num >= 2)
		con_adult.sleep();
	bg.AdultRowToMolokai();
	adult_O_num--;
	adult_M_num++;
	boat ^= 1;
	con_child_M.wake();

	lock.release();
  }

  static void ChildItinerary() {

    bg.initializeChild(); // Required for autograder interface. Must be the
                          // first thing called.
    // DO NOT PUT ANYTHING ABOVE THIS LINE.
	lock.acquire();

	int place = 0;
	while (place != -1){ // Same thing as while(true), but while(true) will make compile error
		
		if (place == 0){
			while ((boat == 1 || child_O_num < 2) && !has_pilot)
				con_child_O.sleep();
			child_O_num--;
			child_M_num++;
			if (!has_pilot){
				bg.ChildRowToMolokai();
				has_pilot = true;
				con_child_O.wake();
			} else {
				bg.ChildRideToMolokai();
				has_pilot = false;
				boat ^= 1;
				if (child_O_num == 0 && adult_O_num == 0)
					finish.wake();
				else 
					con_child_M.wake();
			}

			con_child_M.sleep();
		} else {
			while (boat == 0) con_child_M.sleep();
		
			child_O_num++;
			child_M_num--;
			bg.ChildRowToOahu();
			boat ^= 1;
			
			if (child_O_num < 2)
				con_adult.wake();
			else	
				con_child_O.wake();

			con_child_O.sleep();
		}
		place ^= 1;
	}	

	lock.release();
  }

  static void SampleItinerary() {
    // Please note that this isn't a valid solution (you can't fit
    // all of them on the boat). Please also note that you may not
    // have a single thread calculate a solution and then just play
    // it back at the autograder -- you will be caught.
    System.out.println(
        "\n ***Everyone piles on the boat and goes to Molokai***");
    bg.AdultRowToMolokai();
    bg.ChildRideToMolokai();
    bg.AdultRideToMolokai();
    bg.ChildRideToMolokai();
  }
}

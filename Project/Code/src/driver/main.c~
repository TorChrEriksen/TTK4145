#include "elev.h"
#include <stdio.h>
#include <unistd.h>

typedef int bool;
#define false 0
#define true !false


typedef enum tag_direction { 
    UP = 0, 
    DOWN = 1
} direction_t;

int relocateUp();
int carryOutCommand(int, int);
int callElevator(int, int, elev_button_type_t);
int waitForCommand(int, bool*);
void openDoor(int);
int verifyUserCommand();
int waitForElevatorArrival(int, int);
void moveElevator_OnUserCommand(direction_t);
int verifyFloorPosition(int);
void updateFloorLight(int);
void stopElevator();

int main(){
    // Initialize hardware
    if (!elev_init()) {
        printf(__FILE__ ": Unable to initialize elevator hardware\n");
        return 1;
    }
    
    printf("Press STOP button to stop elevator and exit program.\n");
    
    // User requested the elevator to pick him/her up at a Floor (NOT YET ISSUED A MOVE COMMAND)
    int lastOrderCommand = -1;
    
    // User requested to be elevated to a floor after entering an elevator arriving at his/her floor.
    int command = -1;
    
    // Used if the elevator is requested from another floor while the elevator has arrived at a floor
    // based on a previous call but not been assigned any floor to elevate to by people at that floor.
    bool requestAborted = false;
    
    //last elevator position
    int lastPosition = relocateUp();
      
    while(true) {
        lastPosition = verifyFloorPosition(lastPosition);
        
        if (elev_get_stop_signal()) {
            stopElevator();
            break;
        }
        
        if (lastOrderCommand != -1)
        {
            lastOrderCommand = -1;
            openDoor(lastPosition);
            lastPosition = waitForCommand(lastPosition, &requestAborted);
            if (requestAborted)
            {
                requestAborted = false;
                lastOrderCommand = lastPosition;
            }
            if (lastPosition == -2)
            {
                stopElevator();
                break;
            }
        }
        
        // Requested to be elevated to the same floor from inside the elevator
        if (elev_get_button_signal(BUTTON_COMMAND, lastPosition) == 1)
        {
            printf("You are already on this floor numbnuts!");
            continue;
        }
        
        // TODO: When issuing an order, keep the order in a Fault Tolerant State
        else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 1) )
        {
            lastPosition = callElevator(1, lastPosition, BUTTON_CALL_DOWN);
            lastOrderCommand = 1;
        }
        else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 2) )
        {
            lastPosition = callElevator(2, lastPosition, BUTTON_CALL_DOWN);
            lastOrderCommand = 2;
        }
        else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 3) )
        {
            lastPosition = callElevator(3, lastPosition, BUTTON_CALL_DOWN);
            lastOrderCommand = 3;
        }
        else if ( elev_get_button_signal(BUTTON_CALL_UP, 0) )
        {
            lastPosition = callElevator(0, lastPosition, BUTTON_CALL_UP);
            lastOrderCommand = 0;
        }
        else if ( elev_get_button_signal(BUTTON_CALL_UP, 1) )
        {
            lastPosition = callElevator(1, lastPosition, BUTTON_CALL_UP);
            lastOrderCommand = 1;
        }
        else if ( elev_get_button_signal(BUTTON_CALL_UP, 2) )
        {
            lastPosition = callElevator(2, lastPosition, BUTTON_CALL_UP);
            lastOrderCommand = 2;
        }
        else if (elev_get_button_signal(BUTTON_COMMAND, 0) == 1)
        {
            command = 0;
            lastPosition = carryOutCommand(command, lastPosition);
            command = -1;
        }
        else if (elev_get_button_signal(BUTTON_COMMAND, 1) == 1)
        {
            command = 1;
            lastPosition = carryOutCommand(command, lastPosition);
            command = -1;
        }
        else if (elev_get_button_signal(BUTTON_COMMAND, 2) == 1)
        {
            command = 2;
            lastPosition = carryOutCommand(command, lastPosition);
            command = -1;
        }
        else if (elev_get_button_signal(BUTTON_COMMAND, 3) == 1)
        {
            command = 3;
            lastPosition = carryOutCommand(command, lastPosition);
            command = -1;
        }
        // end Fault Tolerant State
        else
        {
            continue;
        }
    }
       
    return 0;
}

// Puts the elevator at a Floor, given that it is not already above the top floor.
// This is used to avoid having a malfunction when the program starts and the elevator is not located at a floor.
// The elevator cannot be located above the top floor, then it will hit the power off switch...
// TODO: avoid this problem?
int relocateUp()
{
    elev_set_speed(300);
    while(true)
    {
        int position = verifyFloorPosition(-1);
        if (position != -1)
        {
            elev_set_speed(0);
            return position;
        }
    }
}

int callElevator(int originFloor, int lastPos, elev_button_type_t buttonType)
{
    // [4] = floor
    // [3] = button type
    bool lightControl[4][3];
    for (int n = 0; n < 4; n++)
    {
        for (int k = 0; k < 3; k++)
        {
            lightControl[n][k] = false;
        }
    }
    
    if (lightControl[originFloor][buttonType] == false)
    {
        elev_set_button_lamp(buttonType, originFloor, 1);
        lightControl[originFloor][buttonType] = true;
    }
    
    // Go down
    if ( lastPos > originFloor )
    {
        elev_set_speed(-100);
    }
    
    // Go up
    else if (lastPos < originFloor)
    {
        elev_set_speed(100);
    }
    
    // Elevator is currently at the floor where the request is originating from.
    else if (elev_get_floor_sensor_signal() == lastPos)
    {
        if (lightControl[originFloor][buttonType] == false)
        {
            elev_set_button_lamp(buttonType, originFloor, 1);
            lightControl[originFloor][buttonType] = true;
        }
    }
    
    // Go to the first floor below where the elevator is
    else 
    {
        elev_set_speed(-300);
    }
    
    while(true)
    {
        lastPos = verifyFloorPosition(lastPos);
        
        if (elev_get_stop_signal())
        {
            elev_set_speed(0);
            break;
        }
        // We have arrived at the destination floor, stop the elevator!
        if (lastPos == originFloor)
        {
            elev_set_speed(0);
            if (lightControl[originFloor][buttonType] == true)
            {
                sleep(1);
                lightControl[originFloor][buttonType] = false;
                elev_set_button_lamp(buttonType, originFloor, 0);
            }
            break;
        }
    }
    return lastPos;
}

int carryOutCommand(int userCommand, int lastPos)
{
    if (userCommand > lastPos)
    {
        moveElevator_OnUserCommand(UP);
        // TODO: what if elevator stops before arriving!
        return waitForElevatorArrival(userCommand, lastPos);
    }
    else
    {
        moveElevator_OnUserCommand(DOWN);
        // TODO: what if elevator stops before arriving!
        return waitForElevatorArrival(userCommand, lastPos);                     
    }
}

int waitForCommand(int lastPos, bool *abort)
{
    while(true)
    {
        int userCommand = verifyUserCommand();
        if (userCommand != -1)
        {
            elev_set_door_open_lamp(0); //More sofiticated pls
            switch (userCommand)
            {
                case 0:
                {
                    return carryOutCommand(userCommand, lastPos);
                }
                case 1:
                {
                    return carryOutCommand(userCommand, lastPos);
                }
                case 2:
                {
                    return carryOutCommand(userCommand, lastPos);
                }
                case 3: 
                {
                    return carryOutCommand(userCommand, lastPos);
                }
                case N_FLOORS:
                {
                    *abort = true;
                    return callElevator(1, lastPos, BUTTON_CALL_DOWN);
                }
                case (N_FLOORS + 1):
                {
                    *abort = true;
                    return callElevator(2, lastPos, BUTTON_CALL_DOWN);
                }
                case (N_FLOORS + 2):
                {
                    *abort = true;
                    return callElevator(3, lastPos, BUTTON_CALL_DOWN);
                }
                case (N_FLOORS + 3):
                {
                    *abort = true;
                    return callElevator(0, lastPos, BUTTON_CALL_UP);
                }
                case (N_FLOORS + 4):
                {
                    *abort = true;
                    return callElevator(1, lastPos, BUTTON_CALL_UP);
                }
                case (N_FLOORS + 5):
                {
                    *abort = true;
                    return callElevator(2, lastPos, BUTTON_CALL_UP);
                }
                case 10:
                {
                    return -2;
                }
                default:
                {
                    printf("Please request a valid floor");
                    break;
                }
            }
        }
        else {continue;}
    }
}

void openDoor(int lastPos)
{
    //Open door
    elev_set_floor_indicator(lastPos);
    elev_set_door_open_lamp(1);
 }

int verifyUserCommand()
{
    if ( elev_get_button_signal(BUTTON_COMMAND, 0) == 1 )
    {
        return 0;
    }
    else if ( elev_get_button_signal(BUTTON_COMMAND, 1) == 1 )
    {
        return 1;
    }
    else if ( elev_get_button_signal(BUTTON_COMMAND, 2) == 1 )
    {
        return 2;
    }
    else if ( elev_get_button_signal(BUTTON_COMMAND, 3) == 1 )
    {
        return 3;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 1) )
    {
        return 4;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 2) )
    {
        return 5;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_DOWN, 3) )
    {
        return 6;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_UP, 0) )
    {
        return 7;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_UP, 1) )
    {
        return 8;
    }
    else if ( elev_get_button_signal(BUTTON_CALL_UP, 2) )
    {
        return 9;
    }
    else if (elev_get_stop_signal())
    {
        return 10;
    }
    else
    {
        return -1;
    }
}

//TODO: what if the elevator stops! OMG!
int waitForElevatorArrival(int targetFloor, int lastPos)
{
    int dummy = lastPos;
    // Elevator should be moving, wait for the elevator to arrive at target Floor.
    while(true)
    {
        dummy = verifyFloorPosition(dummy);
        if (elev_get_floor_sensor_signal() == targetFloor)
        {
            dummy = verifyFloorPosition(dummy);
            elev_set_speed(0);
            elev_set_door_open_lamp(1);
            sleep(2); // TODO: Just cosmetic
            elev_set_door_open_lamp(0);
            return elev_get_floor_sensor_signal();
        }
    }
}

void moveElevator_OnUserCommand(direction_t direction)
{
    if (direction == UP) {elev_set_speed(100);}
    else {elev_set_speed(-100);}
}

int verifyFloorPosition(int lastKnownPos)
{
    int signal = elev_get_floor_sensor_signal();
    if (signal > -1 && signal < N_FLOORS)
    {
        updateFloorLight(signal);
        return signal;
    }
    return lastKnownPos;
}

void updateFloorLight(int floor)
{
    elev_set_floor_indicator(floor);
}

void stopElevator()
{
    printf("The elevator has come to a conclusion: stopping\n");
    elev_set_speed(0);
}
